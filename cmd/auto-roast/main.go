package main

import (
	"context"
	"os"

	"github.com/calvinmclean/autoroast/controller"
	"github.com/calvinmclean/autoroast/ui"
)

func main() {
	c, err := controller.NewFromEnv()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	ctx := context.Background()

	if os.Getenv("ENABLE_UI") == "true" {
		roasterUI := ui.NewRoasterUI()
		roasterUI.Run(ctx)
		return
	}

	err = c.Run(ctx)
	if err != nil {
		panic(err)
	}
}
