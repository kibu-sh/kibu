package main

import (
	"github.com/discernhq/devx/cmd/devx/cmd"
	"log"
)

func main() {
	cmd, err := cmd.InitCLI()
	if err != nil {
		log.Fatalln(err)
	}

	if err = cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
