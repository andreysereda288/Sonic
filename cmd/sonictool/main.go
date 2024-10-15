package main

import (
	"fmt"
	"os"

	"github.com/Fantom-foundation/go-opera/cmd/sonictool/cmd"
)

func main() {
	if err := cmd.RunSonicTool(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
