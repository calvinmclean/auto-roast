package main

import (
	"context"
	"flag"
	"os"

	"github.com/calvinmclean/autoroast/controller"
	"github.com/calvinmclean/autoroast/ui"
)

func main() {
	var sessionName, probesInput string
	var showUI, debugUI bool
	flag.StringVar(&sessionName, "session", "", "Session name for TWChart")
	flag.StringVar(&probesInput, "probes", "", "Set probe mapping in format \"1=Name,2=Name,...\". Default is 1=Ambient,2=Beans")
	flag.BoolVar(&showUI, "ui", true, "Enable/disable the UI. Default true")
	flag.BoolVar(&debugUI, "debug", false, "Run UI in debug mode with a terminal")
	flag.Parse()

	cfg := controller.NewConfigFromEnv()
	if sessionName != "" {
		cfg.SessionName = sessionName
	}
	if probesInput != "" {
		cfg.ProbesInput = probesInput
	}

	if !showUI {
		runCLI(cfg)
		return
	}

	runUI(cfg, debugUI)
}

func runUI(cfg controller.Config, debug bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	roasterUI := ui.NewRoasterUI()

	roasterUI.Run(ctx, cfg, debug)
	cancel()
}

func runCLI(cfg controller.Config) {
	c, err := controller.New(cfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.Run(context.Background(), os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
