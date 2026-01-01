package main

import (
	"context"

	"github.com/calvinmclean/autoroast/controller"
)

func main() {
	c, err := controller.NewFromEnv()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
