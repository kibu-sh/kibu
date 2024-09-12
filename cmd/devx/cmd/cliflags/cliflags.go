package cliflags

import "github.com/kibu-sh/kibu/pkg/cli"

var (
	Environment = cli.Flag[string]{
		Long:       "environment",
		Short:      "e",
		Default:    "dev",
		Persistent: true,
	}

	GoogleProject = cli.Flag[string]{
		Long:     "google-project",
		Short:    "",
		Default:  "",
		Required: true,
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

	ConfigSyncRecursive = cli.Flag[bool]{
		Long:        "recursive",
		Short:       "r",
		Description: "If true, syncs all files in the source env to the destination.",
		Default:     false,
	}

	ConfigSyncSrcEnv = cli.Flag[string]{
		Long:        "src",
		Short:       "",
		Required:    true,
		Description: "The source environment to copy the config from.",
	}
	ConfigSyncDestEnv = cli.Flag[string]{
		Long:        "dest",
		Short:       "",
		Description: "The destination environment to copy the config to.",
		Required:    true,
	}

	MigrateDir = cli.Flag[string]{
		Long:        "dir",
		Short:       "d",
		Persistent:  true,
		Required:    true,
		AsDirectory: true,
	}

	MigrateDatabaseUrl = cli.Flag[string]{
		Long:        "database-url",
		Short:       "",
		Description: "The DSN connection string for the database",
		Persistent:  true,
		Required:    true,
	}
)
