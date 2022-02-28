package command

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log/level"
	"trellis.tech/kolekti/prome_exporters/agent"
)

func Run(a *agent.Agent) int {

	if err := a.Run(); err != nil {
		level.Error(a.Logger).Log("failed_run_agent", a.Config, "error", err)
		return 3
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGUSR2)
	<-ch
	a.Stop()
	return 0
}
