package watchtasks

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (b *Builder) runCmd(c Command) (err error) {
	flags := os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	logFilePath := filepath.Join(b.tmpDir, fmt.Sprintf("%s.log", c.Cmd))
	logFile, err := os.OpenFile(logFilePath, flags, 0644)
	if err != nil {
		return err
	}

	logPipe := io.MultiWriter(logFile, os.Stdout)

	cmd := exec.CommandContext(b.rootCtx, c.Cmd, c.Args...)
	cmd.Env = os.Environ()
	cmd.WaitDelay = time.Second * 5
	cmd.Stdout = logPipe
	cmd.Stderr = logPipe

	b.log.Info("starting", slog.String("cmd", cmd.String()))
	if err = cmd.Start(); err != nil {
		goto cleanup
	}

	if err = cmd.Wait(); err != nil {
		goto cleanup
	}

cleanup:
	_ = logFile.Close()
	b.log.Info(cmd.String(), "exited with", err)
	return
}
