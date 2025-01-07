package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"regexp"
	"strings"
)

func isValidEmail(email string) bool {
	const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailPattern)
	return re.MatchString(email)
}

// TODO: Improve the logs/out

func AddLoginCmd(rootCmd *cobra.Command, userManager *UserManager) {
	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login to your wowa account",
		RunE: func(cmd *cobra.Command, args []string) error {
			currentEmail, err := userManager.GetUserEmail()
			if err != nil {
				return err
			}
			if currentEmail != "" {
				fmt.Printf("You are already logged in as %s%s%s\n", AnsiBlue, currentEmail, AnsiReset)
				return nil
			}

			fmt.Println(AnsiYellow + ">" + AnsiReset + "  What is your email?")
			reader := bufio.NewReader(os.Stdin)
			rawEmail, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			email := strings.TrimSpace(rawEmail)
			if !isValidEmail(email) {
				return errors.New("invalid email: " + email)
			}

			fmt.Println(AnsiYellow + ">" + AnsiReset + "  What is your password?")
			password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			if len(password) == 0 {
				return errors.New("the password is too short")
			}

			err = userManager.SignIn(email, string(password))
			if err != nil {
				return err
			}

			fmt.Printf("Successfully signed in as %s%s%s!\n", AnsiBlue, email, AnsiReset)
			return nil
		},
	}
	rootCmd.AddCommand(loginCmd)
}
