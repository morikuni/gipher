package main

import (
	"os"

	"github.com/morikuni/gipher/app"
)

func main() {
	a := app.NewApp()
	os.Exit(a.Run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
