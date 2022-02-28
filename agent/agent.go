package agent

import (
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"trellis.tech/kolekti/prome_exporters/conf"
	"trellis.tech/kolekti/prome_exporters/plugins/inputs"
	"trellis.tech/kolekti/prome_exporters/plugins/outputs"
	"trellis.tech/trellis/common.v1/data-structures/queue"
	"trellis.tech/trellis/common.v1/errcode"
)

// Agent runs a set of plugins.
type Agent struct {
	Config *conf.ExporterConfig

	logger log.Logger

	interval time.Duration

	stopChan chan struct{}

	runningInputs []*runningInput
	runningOutput *runningOutput

	metricsChan chan []*dto.MetricFamily

	metricsBuffer queue.Queue
}

type runningInput struct {
	input    *inputs.Input
	logger   log.Logger
	interval time.Duration

	promeRegisterer prometheus.Registerer
	promeGatherer   prometheus.Gatherer

	promeCollector   inputs.InputPrometheusCollector
	metricsCollector inputs.InputMetricsCollector

	stopChan chan struct{}
}

type runningOutput struct {
	output   outputs.Output
	stopChan chan struct{}
}

// NewAgent returns an Agent for the given Config.
func NewAgent(cfg *conf.ExporterConfig, logger log.Logger) (*Agent, error) {
	a := &Agent{
		Config: cfg,
		logger: logger,

		metricsChan: make(chan []*dto.MetricFamily, len(cfg.Inputs)),

		metricsBuffer: queue.New(),
	}
	if err := a.checkConfig(); err != nil {
		return nil, err
	}
	return a, nil
}

func (p *Agent) checkConfig() error {

	if p.Config.MetricBufferLimit == 0 {
		p.Config.MetricBufferLimit = 10000
	}
	if p.Config.MetricBatchSize == 0 {
		p.Config.MetricBatchSize = 10000
	}
	// inputs
	for _, inputConfig := range p.Config.Inputs {

		interval := time.Duration(inputConfig.Interval)
		if interval < time.Second*10 {
			interval = time.Second * 10
		}
		input, err := inputs.GetFactory(inputConfig.Name)
		if err != nil {
			return err
		}

		logger := log.WithPrefix(p.logger, "input", inputConfig.Name)
		opts := []inputs.Option{
			inputs.Logger(logger),
		}

		if inputConfig.Options != nil {
			opts = append(opts, inputs.Config(inputConfig.Options.ToConfig()))
		}

		runningInput := &runningInput{
			input:    input,
			interval: interval,
			logger:   logger,

			stopChan: make(chan struct{}),
		}
		switch input.InputType() {
		case inputs.InputTypePrometheusCollector:
			runningInput.promeCollector, err = input.NewPrometheusCollector(opts...)
			if err != nil {
				return err
			}
			reg := prometheus.NewRegistry()
			runningInput.promeRegisterer = reg
			runningInput.promeGatherer = reg
			runningInput.promeRegisterer.Register(runningInput.promeCollector)
		case inputs.InputTypeMetricsCollector:
			runningInput.metricsCollector, err = input.NewMetricsCollector(opts...)
			if err != nil {
				return err
			}
		}

		p.runningInputs = append(p.runningInputs, runningInput)
	}

	// output
	{
		outFun, err := outputs.GetFactory(p.Config.Output.Name)
		if err != nil {
			return err
		}

		opts := []outputs.Option{
			outputs.Logger(log.WithPrefix(p.logger, "output", p.Config.Output.Name)),
		}

		if p.Config.Output.Options != nil {
			opts = append(opts, outputs.Config(p.Config.Output.Options.ToConfig()))
		}

		output, err := outFun(opts...)
		if err != nil {
			return err
		}
		p.runningOutput = &runningOutput{output: output, stopChan: make(chan struct{})}
	}
	return nil
}

func (p *Agent) runInputs() error {

	gather := func(input *runningInput) ([]*dto.MetricFamily, error) {

		switch input.input.InputType() {
		case inputs.InputTypePrometheusCollector:
			metrics, err := input.promeGatherer.Gather()
			if err != nil {
				level.Error(input.logger).Log("error", err.Error())
				return nil, err
			}
			return metrics, nil
		case inputs.InputTypeMetricsCollector:
			metrics, err := input.metricsCollector.Gather()
			if err != nil {
				level.Error(input.logger).Log("error", err.Error())
				return nil, err
			}
			return metrics, nil
		}
		return nil, nil
	}
	for _, input := range p.runningInputs {
		if _, err := gather(input); err != nil {
			return err
		}
		go func(in *runningInput) {
			ticker := time.NewTicker(in.interval)
			for {
				select {
				case <-ticker.C:
					level.Info(in.logger).Log("msg", "start_gather")
					metrics, err := gather(input)
					if err != nil {
						continue
					}
					level.Info(in.logger).Log("msg", "input_gather_metrics", "length", len(metrics))
					p.metricsChan <- metrics
				case <-time.After(in.interval):
					level.Error(in.logger).Log("error", "timeout to insert metrics to output channel")
				case <-in.stopChan:
					return
				}
			}
		}(input)
	}
	return nil
}

func (p *Agent) runOutputs() error {
	duration := time.Second * 10
	if time.Duration(p.Config.FlushInterval) > time.Second*10 {
		duration = time.Duration(p.Config.FlushInterval)
	}
	ticker := time.NewTicker(duration)
	output := p.runningOutput
	if output == nil {
		return errcode.Newf("nil running output")
	}

	if err := output.output.Connect(); err != nil {
		return err
	}

	go func(runOut *runningOutput) {
		for {
			select {
			case <-ticker.C:

				lenBuffer := p.metricsBuffer.Length()
				for lenBuffer > 0 {

					batch := p.Config.MetricBatchSize
					if lenBuffer <= p.Config.MetricBatchSize {
						batch = lenBuffer
					}

					level.Info(p.logger).Log("msg", "write_output_size", "buffer_length", lenBuffer, "batch_size", batch)

					metricBuffers, ok := p.metricsBuffer.PopMany(batch)
					if !ok {
						// todo log
						break
					}

					var (
						mapMetrics = make(map[string]*dto.MetricFamily)
						names      []string
					)
					for _, buf := range metricBuffers {
						metric, ok := buf.(*dto.MetricFamily)
						if !ok || metric == nil {
							continue
						}
						mf, ok := mapMetrics[metric.GetName()]
						if ok {
							for _, m := range metric.GetMetric() {
								mf.Metric = append(mf.GetMetric(), m)
							}
						} else {
							mf = metric
							names = append(names, metric.GetName())
						}

						mapMetrics[metric.GetName()] = mf
					}
					if len(names) == 0 {
						level.Warn(p.logger).Log("msg", "write_output_failed", "buffer_length", lenBuffer, "error", "at least one metric")
						continue
					}

					var metrics []*dto.MetricFamily
					for _, name := range names {
						mf := mapMetrics[name]
						for i, m := range mf.GetMetric() {
							for k, v := range p.Config.Tags {
								m.Label = append(m.Label, &dto.LabelPair{Name: &k, Value: &v})
							}
							mf.Metric[i] = m
						}
						metrics = append(metrics, mf)
					}

					if err := runOut.output.Write(metrics); err != nil {
						level.Error(p.logger).Log("msg", "write_output_failed", "buffer_length", lenBuffer, "error", err)
						break
					}

					lenBuffer -= p.Config.MetricBatchSize
				}
			case <-runOut.stopChan:
				if err := runOut.output.Close(); err != nil {
					level.Error(p.logger).Log("msg", "failed_stop_output", "error", err)
				}
				return
			}
		}
	}(output)

	return nil
}

func (p *Agent) runMetricsChan() {
	go func() {
		for {
			select {
			case metrics := <-p.metricsChan:
				level.Info(p.logger).Log("read_buffer_from_metric_chan", len(metrics))
				for _, metric := range metrics {
					lenBuffer := p.metricsBuffer.Length()
					if lenBuffer >= p.Config.MetricBufferLimit {
						level.Warn(p.logger).Log("out_of_the_limit_of_buffer", lenBuffer, "limit", p.Config.MetricBufferLimit)
						continue
					}
					p.metricsBuffer.Push(metric)
				}
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *Agent) stopRunningInputs() {
	for _, input := range p.runningInputs {
		if input.stopChan != nil {
			input.stopChan <- struct{}{}
			close(input.stopChan)
		}
		if input.input.InputType() == inputs.InputTypePrometheusCollector {
			input.promeRegisterer.Unregister(input.promeCollector)
		}
	}
}

func (p *Agent) stopRunningOutputs() {
	if p.runningOutput != nil && p.runningOutput.stopChan != nil {
		p.runningOutput.stopChan <- struct{}{}
	}
}

func (p *Agent) Run() error {
	p.runMetricsChan()
	if err := p.runInputs(); err != nil {
		return err
	}
	if err := p.runOutputs(); err != nil {
		return err
	}
	return nil
}

func (p *Agent) Stop() error {
	p.stopRunningInputs()
	if p.stopChan != nil {
		p.stopChan <- struct{}{}
		close(p.stopChan)
	}
	p.stopRunningOutputs()
	return nil
}
