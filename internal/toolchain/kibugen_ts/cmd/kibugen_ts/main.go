package main

import (
	"fmt"
	"os"

	"github.com/kibu-sh/kibu/internal/toolchain/kibugen_ts"
)

func main() {
	code, err := kibugen_ts.Main()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(code)
}
