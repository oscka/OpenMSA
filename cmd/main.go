package main

import (
        "github.com/fatih/color"
        "os"
        "go-project/internal/app"
)

func main() {
        application := app.NewApp()
        defer application.Cleanup()

        // Verify ansible-playbook is installed
        if err := application.InitializeAnsible(); err != nil {
                color.Red("Error: %v", err)
                os.Exit(1)
        }

        // Initialize playbook structure
        if err := application.CopyPlaybookStructure(); err != nil {
                color.Red("Failed to copy playbook structure: %v\n", err)
                os.Exit(1)
        }

        // Initialize server configuration
        if err := application.SetupServerConfig(); err != nil {
                color.Red("Failed to setup server configuration: %v\n", err)
                os.Exit(1)
        }

        // Create hosts files
        if err := application.CreateHostsFiles(); err != nil {
                color.Red("Failed to create hosts files: %v\n", err)
                os.Exit(1)
        }

        // Start the application
        application.SelectAction()
}
