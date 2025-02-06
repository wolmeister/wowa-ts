package spinny

import (
	"time"
)

func main() {
	manager := NewManager()
	manager.Start()
	defer manager.Stop()

	spinner1 := manager.NewSpinner("Loading task")
	spinner2 := manager.NewSpinner("Loading task")
	spinner3 := manager.NewSpinner("Loading task")
	spinner4 := manager.NewSpinner("Loading task")

	time.Sleep(1 * time.Second)
	spinner1.Succeed("")

	time.Sleep(1 * time.Second)
	spinner2.Warn("")

	time.Sleep(1 * time.Second)
	spinner3.Info("")

	time.Sleep(1 * time.Second)
	spinner4.Fail("")
}
