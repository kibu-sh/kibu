package watchtasks

func NewGenerateStep(arg string) Command {
	return Command{
		Cmd:  "go",
		Args: []string{"generate", arg},
	}
}

func (b *Builder) runBuildCmdWithSignal(done chan error) {
	_ = b.topic.Publish(b.rootCtx, Event{
		Type: EventTypeRebuilding,
	})

	err := b.runCmd(NewGenerateStep("./..."))
	if err == nil {
		b.restart <- struct{}{}
	}
	done <- err
}

func (b *Builder) runBuildLoop() {
	done := make(chan error)
	go b.runBuildCmdWithSignal(done)

	for {
		select {
		case <-b.rootCtx.Done():
			return
		case <-b.rebuild:
			select {
			case <-done:
				go b.runBuildCmdWithSignal(done)
			default:
				b.log.Debug("dropping rebuild signal, previous build is still running")
			}
		}
	}
}
