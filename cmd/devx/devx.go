package main

import (
	"github.com/discernhq/devx/cmd/devx/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	cmd, err := cmd.InitCLI()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize CLI")
	}

	if err = cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute CLI")
	}
}
