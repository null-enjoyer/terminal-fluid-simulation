package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/null-enjoyer/terminal-fluid-simulation/app"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	configPath := flag.String("config", "", "Path to the settings file (optional)")
	help := flag.Bool("help", false, "Show this help message")
	showVersion := flag.Bool("version", false, "Show application version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("Terminal Fluid Simulation\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Commit:  %s\n", Commit)
		fmt.Printf("Date:    %s\n", Date)
		os.Exit(0)
	}

	application := app.New(*configPath, Version)
	defer application.Screen.Fini()
	application.Run()
}
