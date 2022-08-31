package config

// ExternalConfig will hold the configurations for external tools, such as Explorer or Elastic Search
type ExternalConfig struct {
	ElasticSearchConnector ElasticSearchConfig
	EventNotifierConnector EventNotifierConfig
	CovalentConnector      CovalentConfig
	WebSocketsConnector    WebSocketsDriverConfig
}

// ElasticSearchConfig will hold the configuration for the elastic search
type ElasticSearchConfig struct {
	Enabled                   bool
	IndexerCacheSize          int
	BulkRequestMaxSizeInBytes int
	URL                       string
	UseKibana                 bool
	Username                  string
	Password                  string
	EnabledIndexes            []string
}

// EventNotifierConfig will hold the configuration for the events notifier driver
type EventNotifierConfig struct {
	Enabled          bool
	UseAuthorization bool
	ProxyUrl         string
	Username         string
	Password         string
}

// CovalentConfig will hold the configurations for covalent indexer
type CovalentConfig struct {
	Enabled              bool
	URL                  string
	RouteSendData        string
	RouteAcknowledgeData string
}

// WebSocketsDriverConfig will hold the configuration for web sockets driver
type WebSocketsDriverConfig struct {
	Enabled         bool
	WithAcknowledge bool
	URL             string
	MarshallerType  string
}
