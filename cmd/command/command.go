/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

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
