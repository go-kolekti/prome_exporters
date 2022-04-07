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

package main

import (
	"os"

	"trellis.tech/kolekti/prome_exporters/agent"
	"trellis.tech/kolekti/prome_exporters/cmd/command"
	"trellis.tech/kolekti/prome_exporters/cmd/server"
	"trellis.tech/kolekti/prome_exporters/conf"

	"github.com/go-kit/log/level"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	promlog "trellis.tech/trellis/common.v1/logger/prometheus"
	flag "trellis.tech/trellis/common.v1/logger/prometheus/flag"

	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/all"
	_ "trellis.tech/kolekti/prome_exporters/plugins/outputs/all"
	_ "trellis.tech/kolekti/prome_exporters/plugins/serializers/all"
)

func main() {
	os.Exit(run())
}

var (
	promlogConfig = &promlog.Config{}
)

func run() int {
	var (
		cfgFile = kingpin.Flag("config.file", "Exporters configuration file name.").Default("exporters.yaml").String()
	)
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.CommandLine.UsageWriter(os.Stdout)

	kingpin.Version(version.Print("prome_exporters"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	ec, err := conf.GetConfigWithFile(*cfgFile)
	if err != nil {
		level.Error(logger).Log("failed_read_config", cfgFile, "error", err)
		return 1
	}

	a, err := agent.NewAgent(ec, logger)
	if err != nil {
		level.Error(logger).Log("failed_new_agent", ec, "error", err)
		return 2
	}

	switch ec.Exporter.CommandType {
	case 0:
		return command.Run(a)
	case 1:
		return server.Run(a)
	default:
		return 10
	}
}
