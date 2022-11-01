package main

import (
	"github.com/discernhq/devx/cmd/devx/cmd"
	"github.com/rs/zerolog"
	"os"
)

var log = zerolog.New(os.Stderr).With().Timestamp().Logger()

func main() {
	root, err := cmd.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize")
	}

	if err = root.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute")
	}
}
