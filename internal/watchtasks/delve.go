package watchtasks

import (
	"context"
	"fmt"
	"github.com/go-delve/delve/service/rpc2"
	"github.com/pkg/errors"
	"net"
	"path/filepath"
	"time"
)

const (
	DefaultDebugServerAddr = "127.0.0.1:2345"
)

type DebugServerCommandParams struct {
	// PkgDir is the directory of the package to build
	PkgDir string

	// BuildOutDir is the directory where the build output is placed
	BuildOutDir string

	// DebugServerAddr is the address of the debug server
	DebugServerAddr string
}

func NewDebugServerCommand(params DebugServerCommandParams) Command {
	return Command{
		Cmd: "dlv",
		Args: []string{
			"debug",
			"--headless",
			"--continue",
			"--api-version=2",
			"--accept-multiclient",
			debugServerListenFlag(params.DebugServerAddr),
			debugBinaryOutputFlag(params.BuildOutDir),
			params.PkgDir,
		},
	}
}

func debugServerListenFlag(addr string) string {
	return fmt.Sprintf("--listen=%s", addr)
}

func debugBinaryOutputFlag(output string) string {
	return fmt.Sprintf("--output=%s", filepath.Join(output, "__debug"))
}

func (b *Builder) runDebugServer(started chan struct{}) {
	go func() {
		coefficient := 2
		delay := time.Second * 1
		maxDelay := time.Minute * 5

	retry:
		conn, err := net.DialTimeout("tcp", DefaultDebugServerAddr, time.Second*5)
		if err != nil {
			b.log.Debug("waiting for debug server", "error", err)
			time.Sleep(delay)
			delay *= time.Duration(coefficient)
			if delay >= maxDelay {
				delay = maxDelay
			}
			goto retry
		}

		_ = conn.Close()
		b.debuggerListening <- struct{}{}

		_ = b.topic.Publish(b.rootCtx, Event{
			Type: EventTypeDebuggerListening,
		})
	}()

	for {
		err := b.runCmd(NewDebugServerCommand(DebugServerCommandParams{
			PkgDir:          b.appDir,
			BuildOutDir:     b.tmpDir,
			DebugServerAddr: DefaultDebugServerAddr,
		}))
		if err != nil {
			b.log.Debug("debug server exited", "error", err)
		}
		if errors.Is(err, context.Canceled) {
			return
		}
	}

}

func newRPCClient(addr string, timeout time.Duration) (*rpc2.RPCClient, net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, nil, err
	}

	client := rpc2.NewClientFromConn(conn)
	return client, conn, err
}

func (b *Builder) runDebugLoop() {
	started := make(chan struct{})
	<-b.restart

	go b.runDebugServer(started)
	<-b.debuggerListening

	for {
		select {
		case <-b.rootCtx.Done():
			b.log.Debug("shutting down debug server")
			client, _, err := newRPCClient(DefaultDebugServerAddr, time.Second*5)
			if err != nil {
				b.log.Debug("failed to connect to debug server", "error", err)
				return
			}

			err = client.Detach(true)
			if err != nil {
				b.log.Debug("failed to shutdown debug server", "error", err)
				return
			}
		case <-b.restart:
			b.log.Debug("restarting debug server")
			client, _, err := newRPCClient(DefaultDebugServerAddr, time.Second*5)
			state, _ := client.GetStateNonBlocking()
			if state != nil && state.Running {
				_, err := client.Halt()
				if err != nil {
					b.log.Debug("failed to halt debug server", "error", err)
					continue
				}
			}

			_, err = client.Restart(true)
			if err != nil {
				b.log.Debug("failed to restart debug server", "error", err)
				continue
			}

			_ = client.Disconnect(true)
			_ = b.topic.Publish(b.rootCtx, Event{
				Type: EventTypeDebuggerRestarted,
			})
			b.log.Debug("successfully restarted debug server")
		}
	}
}
