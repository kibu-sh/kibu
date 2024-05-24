package main

import (
	"github.com/discernhq/devx/internal/buffchan"
	"github.com/discernhq/devx/internal/fswatch"
	"github.com/discernhq/devx/internal/watchtasks"
	"github.com/discernhq/devx/internal/watchtasks/watchtui"
	"github.com/discernhq/devx/pkg/appcontext"
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
