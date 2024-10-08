package appcontext

import (
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
	"sync"
)

var cache context.Context
var cancel context.CancelFunc
var doOnce sync.Once

// Context returns a static context that reacts to termination signals of the
// running process. Useful in CLI tools.
func Context() context.Context {
	doOnce.Do(func() {
		signals := make(chan os.Signal, 2048)
		signal.Notify(signals, []os.Signal{unix.SIGTERM, unix.SIGINT}...)

		const exitLimit = 3
		retries := 0

		cache, cancel = context.WithCancel(context.Background())

		go func() {
			for {
				<-signals
				cancel()
				retries++
				if retries >= exitLimit {
					logrus.Errorf("got %d SIGTERM/SIGINTs, forcing shutdown", retries)
					os.Exit(1)
				}
			}
		}()
	})
	return cache
}

func UpdateCache(ctx context.Context) {
	cache = ctx
}

func Dangerous__CancelEscapeHatch() {
	if cancel != nil {
		cancel()
	}
}
