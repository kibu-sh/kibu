package cliflags

import "github.com/discernhq/devx/pkg/cli"

var (
	Environment = cli.Flag[string]{
		Long:    "environment",
		Short:   "e",
		Default: "dev",
		Global:  true,
	}

	Debug = cli.Flag[bool]{
		Long:    "debug",
		Short:   "",
		Default: false,
		Global:  true,
	}

	// Config Flags

	ConfigSetFromFile = cli.Flag[string]{
		Long:       "from-file",
		Short:      "",
		Default:    "",
		AsFilename: true,
	}

	ConfigSetFromLiteral = cli.Flag[[]string]{
		Long:  "from-literal",
		Short: "",
	}

	ConfigSetFromEnvFile = cli.Flag[string]{
		Long:       "from-env-file",
		Short:      "",
		Default:    "",
		AsFilename: true,
	}
)
