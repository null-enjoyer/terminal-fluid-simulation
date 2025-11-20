package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/null-enjoyer/terminal-fluid-simulation/app"
)

func main() {
	configPath := flag.String("config", "", "Path to the settings file (optional)")
	help := flag.Bool("help", false, "Show this help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	application := app.New(*configPath)
	defer application.Screen.Fini()
	application.Run()
}
