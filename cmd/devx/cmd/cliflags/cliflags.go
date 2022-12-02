package cliflags

import "github.com/discernhq/devx/pkg/clioptions"

var (
	Environment = clioptions.Flag[string]{
		Long:    "environment",
		Short:   "e",
		Default: "dev",
		Global:  true,
	}

	Debug = clioptions.Flag[bool]{
		Long:    "debug",
		Short:   "",
		Default: false,
		Global:  true,
	}

	// Config Flags

	ConfigSetFromFile = clioptions.Flag[string]{
		Long:       "from-file",
		Short:      "",
		Default:    "",
		AsFilename: true,
	}

	ConfigSetFromLiteral = clioptions.Flag[[]string]{
		Long:  "from-literal",
		Short: "",
	}

	ConfigSetFromEnvFile = clioptions.Flag[string]{
		Long:       "from-env-file",
		Short:      "",
		Default:    "",
		AsFilename: true,
	}
)
