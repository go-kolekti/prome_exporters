# prome_exporters
A framework for exporters to collect prometheus metrics


## Server Mode (command_type)

> make GOLANGCI_LINT_OPTS="-D errorcheck"

* 1 : run command
* 2 : run server 

```yaml
exporter:
  command_type : 1 

  ### supported Prometheus Blackbox_exporter
  blackbox_probe:
    open: false # command_type = 1 & open = true
```

## input

> input construct function

* func(...inputs.Option) (prometheus.Collector, error)
* func(...inputs.Option) (inputs.InputMetricsCollector, error)

### Feature

* Prometheus NodeExporter (prometheus_node_exporter)
* Supported HTTP GET From API Server, supported parsers: prometheus, jmx, opentsdb (http)
* Zookeeper TCP: mntr (zookeeper)

## output

NewOutputFactory = func(opts ...outputs.Option) (outputs.Output, error)
outputs.RegisterFactory("name", NewOutputFactory)

```go
type Output interface {
    plugins.PluginDescriber

	// Connect to the Output; connect is only called once when the plugin starts
	Connect() error
	// Close any connections to the Output. Close is called once when the output
	// is shutting down. Close will not be called until defaults writes have finished,
	// and Write() will not be called once Close() has been, so locking is not
	// necessary.
	Close() error
	// Write takes in group of points to be written to the Output
	Write(metrics []*dto.MetricFamily) error
}
```

## Parsers

> parsers metrics bytes to map[string]*dto.MetricFamily

### Feature

* Supported Java JMX HTTP Metric (jmx)
* Supported Prometheus HTTP Metric (prometheus)
* Supported OpenTSDB HTTP Metric (opentsdb)

[config sample](exporters_sample.yaml)

## todo

* Output: Metrics To Kafka