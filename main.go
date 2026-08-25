package main

import (
	"fmt"
	"os"

	"binary-parser/internal/cli"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "serve" {
		runServe(os.Args[2:])
		return
	}
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
