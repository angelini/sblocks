package main

import (
	"os"

	"github.com/angelini/sblocks/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
