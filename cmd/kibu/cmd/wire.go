//go:build wireinject
// +build wireinject

package cmd

import "github.com/google/wire"

func InitCLI() (RootCmd, error) {
	wire.Build(WireSet)
	return RootCmd{}, nil
}
