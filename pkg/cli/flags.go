package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Flag[T any] struct {
	// Long is the long name of the flag. "environment" would be --environment
	Long string

	// Short is the short name of the flag. (-e)
	Short string

	// Default is the default value for the flag.
	Default T

	// Required marks a flag as required
	Required bool

	// Persistent marks a flag as global
	// This will persist the flag to all subcommands
	Persistent bool

	// Description a brief explanation of the flag
	Description string

	// AsFilename limit shell completion to only files
	AsFilename bool

	// AsDirectory limit shell completion to only directories
	AsDirectory bool

	// value is the value of the flag only works after flag parsing
	value *T
}

func (f *Flag[T]) Value() T {
	if f.value == nil {
		return f.Default
	}
	return *f.value
}

func (f *Flag[T]) BindToCommand(cmd *cobra.Command) error {
	f.value = new(T)

	var flags = cmd.Flags()

	if f.Persistent {
		flags = cmd.PersistentFlags()
	}

	defaultValue := any(f.Default)

	switch bind := any(f.value).(type) {
	case *string:
		flags.StringVarP(bind, f.Long, f.Short, defaultValue.(string), f.Description)
		return nil
	case *bool:
		flags.BoolVarP(bind, f.Long, f.Short, defaultValue.(bool), f.Description)
		return nil
	case *[]string:
		flags.StringSliceVarP(bind, f.Long, f.Short, defaultValue.([]string), f.Description)
	default:
		return errors.Errorf("unsupported cli option type %T", defaultValue)
	}

	if f.Required {
		if err := cmd.MarkFlagRequired(f.Long); err != nil {
			return err
		}
	}

	if f.AsFilename {
		if err := cmd.MarkFlagDirname(f.Long); err != nil {
			return err
		}
	}

	return nil
}
