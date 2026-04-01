package main

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/bootstrap"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "Catalog-microservice",
	Short:            "Catalog-microservice based on clean architecture",
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
