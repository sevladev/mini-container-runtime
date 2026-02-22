package main

import (
	"fmt"
	"os"

	"github.com/sevladev/minic/internal/cli"
	"github.com/sevladev/minic/internal/container"
)

func main() {
	// TODO: /proc/self/exe init — intercepted before cobra
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := container.Init(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "init error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
