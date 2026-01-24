package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"mklink/ui"
	"mklink/utils"
)

func main() {
	// Enable ANSI colors
	_ = utils.EnableVirtualTerminalProcessing()

	if !utils.IsAdmin() {
		fmt.Println("Requesting Administrative Privileges...")
		if err := utils.RunAsAdmin(); err != nil {
			fmt.Printf("Failed to elevate: %v\n", err)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(1)
		}
		os.Exit(0)
	}

	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
