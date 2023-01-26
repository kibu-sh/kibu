package cliflags

import "github.com/discernhq/devx/pkg/cli"

var (
	Environment = cli.Flag[string]{
		Long:       "environment",
		Short:      "e",
		Default:    "dev",
		Persistent: true,
	}

	Debug = cli.Flag[bool]{
		Long:       "debug",
		Short:      "",
		Default:    false,
		Persistent: true,
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

	MigrateDir = cli.Flag[string]{
		Long:        "dir",
		Short:       "d",
		Persistent:  true,
		Required:    true,
		AsDirectory: true,
	}

	MigrateDatabaseUrl = cli.Flag[string]{
		Long:       "database-url",
		Short:      "db",
		Persistent: true,
		Required:   true,
	}
)
