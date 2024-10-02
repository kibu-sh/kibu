package main

import (
	"fmt"
	"github.com/kibu-sh/kibu/internal/toolchain/kibuwire"
	"os"
)

func main() {
	code, err := kibuwire.Main()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(code)
}
