package main

import (
	"github.com/kibu-sh/kibu/internal/buffchan"
	"github.com/kibu-sh/kibu/internal/fswatch"
	"github.com/kibu-sh/kibu/internal/watchtasks"
	"github.com/kibu-sh/kibu/internal/watchtasks/watchtui"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"os"
	"path/filepath"
	"time"
)

func main() {
	ctx := appcontext.Context()
	root := "/Users/jqualls/projects/github.com/discernhq/discern"
	appDir := filepath.Join(root, "src/backend/cmd/server")
	err := os.Chdir(root)
	if err != nil {
		panic(err)
	}

	w, err := fswatch.NewRecursiveWatcher(ctx, root,
		fswatch.WithDetectGitRoot(),
	)
	if err != nil {
		return
	}
	w.Start()
	defer w.Stop(time.Second * 5)

	buffFSEvents := buffchan.NewWindowBufferedChannel(ctx, buffchan.WindowBufferChannelOptions[fswatch.Event]{
		Window:   time.Second * 2,
		Input:    w.Events(),
		Debounce: fswatch.DebounceOnGitOp,
	})

	b, err := watchtasks.NewBuilder(ctx,
		watchtasks.WithFSEvents(buffFSEvents),
		watchtasks.WithRootDir(root),
		watchtasks.WithAppDir(appDir),
	)
	if err != nil {
		return
	}

	b.Start()
	defer b.Stop(time.Second * 5)

	events, err := b.Subscribe(ctx)
	if err != nil {
		return
	}

	go func() {
		err = watchtui.Start(ctx, events)
		appcontext.Dangerous__CancelEscapeHatch()
		if err != nil {
			return
		}
	}()

	<-ctx.Done()
}
