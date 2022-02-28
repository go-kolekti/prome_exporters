package plugins

// PluginDescriber contains the functions all plugins must implement to describe
// themselves to Telegraf. Note that all plugins may define a logger that is
// not part of the interface, but will receive an injected logger if it's set.
// eg: Log telegraf.Logger `toml:"-"`
type PluginDescriber interface {
	// SampleConfig returns the default configuration of the Processor
	SampleConfig() string

	// Description returns a one-sentence description on the Processor
	Description() string
}
