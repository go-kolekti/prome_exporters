package all

import (
	// prometheus_go
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/prometheus_go"
	// prometheus_node_exporter
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/promethues_node_exporter"
	// prometheus_node_exporter
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/prometheus_process"
	// metric_http
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/metric_http"
)
