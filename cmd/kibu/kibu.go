package main

import (
	"fmt"
	"github.com/kibu-sh/kibu/cmd/kibu/cmd"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

func panicRecovery(code *int) {
	if r := recover(); r != nil {
		fmt.Printf("panic recovery %v\n", r)
		*code = 1
	}
}

func execute() (code int) {
	var err error
	var c cmd.RootCmd
	defer panicRecovery(&code)

	defer func() {
		if err != nil {
			fmt.Printf("error: %v\n", err)
			code = 1
		}
	}()

	c, err = cmd.InitCLI()
	if err != nil {
		return
	}

	if err = c.Execute(); err != nil {
		return
	}
	return
}

func isProfilerEnabled() bool {
	return strings.TrimSpace(os.Getenv("KIBU_PROFILER_ENABLED")) == "true"
}

func profileCPU(enabled bool) func() {
	if !enabled {
		return func() {}
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cpuFile := filepath.Join(cwd, "cpu.prof")

	f, err := os.OpenFile(cpuFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}

	return pprof.StopCPUProfile
}

func profileHeap(enabled bool) {
	if !enabled {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	heapFile := filepath.Join(cwd, "heap.prof")

	f, err := os.OpenFile(heapFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
	return
}

func main() {
	stop := profileCPU(isProfilerEnabled())
	code := execute()

	stop()
	profileHeap(isProfilerEnabled())
	os.Exit(code)
}
