package all

import (
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/jmx"
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/opentsdb"
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/prometheus"
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/promethues_node_exporter"
	_ "trellis.tech/kolekti/prome_exporters/plugins/inputs/zookeeper"
)
