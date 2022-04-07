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
