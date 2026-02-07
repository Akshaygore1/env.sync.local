package main

import (
	"os"

	"envsync/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
