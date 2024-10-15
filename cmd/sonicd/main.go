package main

import (
	"fmt"
	"os"

	"github.com/Fantom-foundation/go-opera/cmd/sonicd/cmd"
)

func main() {
	if err := cmd.RunSonicd(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
