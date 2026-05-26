package main

import (
	"fmt"
	"os"
)

func main() {
	// Execute the root command from your package context

	rootCmd.AddCommand(initCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
