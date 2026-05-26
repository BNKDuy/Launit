package main

import (
	"fmt"
	"os"
)

func main() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
