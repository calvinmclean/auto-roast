package main

import (
	"context"
	"io"
	"os"

	"github.com/calvinmclean/autoroast/controller"
	"github.com/calvinmclean/autoroast/ui"
)

func main() {
	if os.Getenv("ENABLE_UI") == "true" {
		runUI()
		return
	}

	runCLI()
}

func runUI() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := controller.NewFromEnv()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	r, w := io.Pipe()

	// read from Stdin also
	go func() {
		defer w.Close()
		io.Copy(w, os.Stdin)
	}()

	roasterUI := ui.NewRoasterUI()

	go func() {
		err = c.Run(ctx, r, io.MultiWriter(os.Stdout, roasterUI))
		if err != nil {
			panic(err)
		}
	}()

	roasterUI.Run(ctx, w)
	cancel()
}

func runCLI() {
	c, err := controller.NewFromEnv()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.Run(context.Background(), os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
