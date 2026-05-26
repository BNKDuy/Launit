package main

import (
	"fmt"
	"orchestrator/internal/config"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Launitfile for your project",
	Run: func(cmd *cobra.Command, args []string) {
		filename := "Launitfile"

		// Check if the file already exists
		if _, err := os.Stat(filename); err == nil {
			fmt.Println("Launitfile already exists in this directory!")
			return
		}

		starterContent, err := config.NewYamlCliConfig().GetTemplateBytes()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if err := os.WriteFile(filename, starterContent, 0644); err != nil {
			fmt.Printf("Error creating Launitfile: %v\n", err)
			return
		}

		fmt.Println("Launitfile Created!")
	},
}
