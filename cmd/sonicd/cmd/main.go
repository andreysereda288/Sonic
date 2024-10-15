package cmd

import (
	"os"
)

func RunSonicd() error {
	initApp()
	initAppHelp()
	return app.Run(os.Args)
}
