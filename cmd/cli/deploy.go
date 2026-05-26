package main

import (
	"fmt"
	"orchestrator/internal/api"
	"orchestrator/internal/config"
	"orchestrator/internal/packager"
	"os"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Compile and deply your appplication to a serverless function",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewYamlCliConfig().Load()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		packager, err := packager.New(cfg.App.Runtime)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		zip, err := packager.BuildAndPackage(cfg.App.Path)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer os.Remove(zip)

		httpClient := api.NewClient("http://localhost:8080")
		presignUrl, err := httpClient.GetPresignedURL(cfg.App.Name)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if err := httpClient.UploadZip(presignUrl, zip); err != nil {
			fmt.Println(err.Error())
			return
		}

		url, err := httpClient.SendCreateRequest(cfg.App.Name, cfg.App.Runtime, cfg.App.Size, cfg.App.Timeout)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println("✨ Success! Your app is live at: " + url)
	},
}
