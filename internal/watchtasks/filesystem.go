package watchtasks

import (
	"github.com/kibu-sh/kibu/internal/fswatch"
	"slices"
)

func (b *Builder) runEventLoop() {
	var err error
	b.finished = make(chan error)
	go b.runBuildLoop()
	go b.runDebugLoop()

	for {
		select {
		case <-b.rootCtx.Done():
			err = b.rootCtx.Err()
			goto cleanup
		case fse, ok := <-b.fsEvents:
			if !ok {
				goto cleanup
			}

			if slices.ContainsFunc(fse, fswatch.IsGeneratedFile) {
				continue
			}

			b.rebuild <- struct{}{}
		}
	}

cleanup:
	b.finished <- err
	close(b.finished)
	return
}
