package main

import (
	"fmt"
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2"
	"os"
)

func main() {
	code, err := kibugenv2.Main()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(code)
}
