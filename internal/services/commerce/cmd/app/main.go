package main

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/bootstrap"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "commerce-microservice",
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
