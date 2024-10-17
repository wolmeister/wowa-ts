package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database structs
type User struct {
	Id        string    `gorm:"primarykey;not null"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time
}

type GameVersion string

const (
	Retail  GameVersion = "retail"
	Classic GameVersion = "classic"
)

type Provider string

const (
	Curse Provider = "curse"
)

type Addon struct {
	Id          string      `gorm:"primarykey;not null" json:"id"`
	UserId      string      `gorm:"uniqueIndex:idx_unique_user_addon;not null" json:"user_id"`
	GameVersion GameVersion `gorm:"uniqueIndex:idx_unique_user_addon;not null" json:"game_version"`
	Slug        string      `gorm:"uniqueIndex:idx_unique_user_addon;not null" json:"slug"`
	Name        string      `gorm:"not null" json:"name"`
	Author      string      `gorm:"not null" json:"author"`
	Provider    Provider    `gorm:"not null" json:"provider"`
	ExternalId  string      `gorm:"not null" json:"external_id"`
	Url         string      `gorm:"not null" json:"url"`
	CreatedAt   time.Time   `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"not null" json:"updated_at"`
	User        User        `json:"-"`
}

// Request structs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4"`
}

type AddAdonnRequest struct {
	GameVersion GameVersion `json:"game_version" validate:"required,oneof=retail classic"`
	Slug        string      `json:"slug" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	Author      string      `json:"author" validate:"required"`
	Provider    Provider    `json:"provider" validate:"required,oneof=curse"`
	ExternalId  string      `json:"external_id" validate:"required"`
	Url         string      `json:"url" validate:"required,url"`
}

func loginHandler(db *gorm.DB, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginRequest LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if err := validate.Struct(loginRequest); err != nil {
			http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
			return
		}

		var user User
		result := db.First(&user, "email = ?", loginRequest.Email)
		if result.Error != nil {
			if result.Error != gorm.ErrRecordNotFound {
				log.Println("Error", result.Error)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			log.Println("Creating new user")
			userId := "user_" + uuid.New().String()
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(loginRequest.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			user = User{Id: userId, Email: loginRequest.Email, Password: string(hashedPassword)}
			result := db.Create(&user)
			if result.Error != nil {
				log.Println("Error", result.Error)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("User already exists")
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
			if err != nil {
				http.Error(w, "Invalid password", http.StatusBadRequest)
				return
			}
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.Id,
			"exp": time.Now().AddDate(1, 0, 0).Unix(),
		})
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
		if err != nil {
			log.Println("Error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(tokenString))
	}
}

func getUserIdFromRequest(w http.ResponseWriter, r *http.Request) string {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(tokenString) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return ""
	}
	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is what you expect
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})
	if err != nil || !parsedToken.Valid {
		log.Println("Error", err, parsedToken.Valid)
		http.Error(w, "Unauthorized (invalid token)", http.StatusUnauthorized)
		return ""
	}
	userId, err := parsedToken.Claims.GetSubject()
	if err != nil {
		log.Println("Error", err)
		http.Error(w, "Unauthorized (invalid token)", http.StatusUnauthorized)
		return ""
	}
	return userId
}

func createAddonHandler(db *gorm.DB, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(w, r)
		if len(userId) == 0 {
			return
		}

		var addAddonRequest AddAdonnRequest
		if err := json.NewDecoder(r.Body).Decode(&addAddonRequest); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if err := validate.Struct(addAddonRequest); err != nil {
			http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
			return
		}

		addonId := "addon_" + uuid.New().String()
		addon := Addon{
			Id:          addonId,
			UserId:      userId,
			GameVersion: addAddonRequest.GameVersion,
			Slug:        addAddonRequest.Slug,
			Name:        addAddonRequest.Name,
			Author:      addAddonRequest.Author,
			Provider:    addAddonRequest.Provider,
			ExternalId:  addAddonRequest.ExternalId,
			Url:         addAddonRequest.Url,
		}
		result := db.Create(&addon)
		if result.Error != nil {
			if result.Error == gorm.ErrDuplicatedKey {
				// TODO: Check if everything is the same and not throw?
				http.Error(w, "Addon already exists", http.StatusConflict)
				return
			}
			log.Println("Error", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err := json.NewEncoder(w).Encode(&addon)
		if err != nil {
			log.Println("Error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func getAddonsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(w, r)
		if len(userId) == 0 {
			return
		}

		var addons []Addon
		result := db.Where("user_id = ?", userId).Find(&addons)
		if result.Error != nil {
			log.Println("Error", result.Error)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		log.Println(addons)
		err := json.NewEncoder(w).Encode(&addons)
		if err != nil {
			log.Println("Error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

}

func getAddonHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(w, r)
		if len(userId) == 0 {
			return
		}

		vars := mux.Vars(r)
		gameVersion := vars["game_version"]
		slug := vars["slug"]

		var addon Addon
		result := db.Where("user_id = ? AND game_version = ? AND slug = ?", userId, gameVersion, slug).First(&addon)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				http.Error(w, "Not found error", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err := json.NewEncoder(w).Encode(&addon)
		if err != nil {
			log.Println("Error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func deleteAddonHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := getUserIdFromRequest(w, r)
		if len(userId) == 0 {
			return
		}

		vars := mux.Vars(r)
		gameVersion := vars["game_version"]
		slug := vars["slug"]

		result := db.Where("user_id = ? AND game_version = ? AND slug = ?", userId, gameVersion, slug).Delete(&Addon{})
		if result.Error != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if result.RowsAffected == 0 {
			http.Error(w, "Not found error", http.StatusNotFound)
			return
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}

func main() {
	// Setup database
	dsn := os.Getenv("PG_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// DryRun: true
		TranslateError: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&User{}, &Addon{})
	if err != nil {
		log.Fatal(err)
	}

	// Create the json validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Start http server
	r := mux.NewRouter()
	r.HandleFunc("/addons/{game_version}/{slug}", getAddonHandler(db)).Methods("GET")
	r.HandleFunc("/addons/{game_version}/{slug}", deleteAddonHandler(db)).Methods("DELETE")
	r.HandleFunc("/addons", getAddonsHandler(db)).Methods("GET")
	r.HandleFunc("/addons", createAddonHandler(db, validate)).Methods("POST")
	r.HandleFunc("/login", loginHandler(db, validate)).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	http.Handle("/", r)

	log.Println("Server started at :8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
