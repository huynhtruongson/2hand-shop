package main

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/bootstrap"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "identity-microservice",
	Short:            "identity-microservice based on clean architecture",
	Long:             `This is a command runner or cli for api architecture in golang.`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		bootstrap.NewApp().Run()
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
