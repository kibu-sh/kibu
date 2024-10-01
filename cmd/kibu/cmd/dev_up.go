package cmd

import (
	"fmt"
	"github.com/kibu-sh/kibu/internal/buffchan"
	"github.com/kibu-sh/kibu/internal/fswatch"
	"github.com/kibu-sh/kibu/internal/watchtasks"
	"github.com/kibu-sh/kibu/pkg/appcontext"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

type DevUpCmd struct {
	*cobra.Command
}

type NewDevUpCmdParams struct{}

func NewDevUpCmd(params NewDevUpCmdParams) (cmd DevUpCmd) {
	cmd.Command = &cobra.Command{
		Use:   "up",
		Short: "up",
		Long:  `up`,
		RunE:  newDevUpRunE(params),
	}
	return
}

func newDevUpRunE(params NewDevUpCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		ctx := appcontext.Context()
		root, err := os.Getwd()
		if err != nil {
			return
		}

		fmt.Println("watching", root)

		// TODO: make this configurable
		appDir := filepath.Join(root, "src/backend/cmd/server")
		err = os.Chdir(root)
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

		buffFSEvents := buffchan.NewWindowBufferedChannel(ctx,
			buffchan.WindowBufferChannelOptions[fswatch.Event]{
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

		//go func() {
		//	err = watchtui.Start(ctx, events)
		//	appcontext.Dangerous__CancelEscapeHatch()
		//	if err != nil {
		//		return
		//	}
		//}()

		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case e := <-events.Channel():
				fmt.Println(e.Type.String())
			}
		}
	}
}
