package entities

type Database struct {
	Driver           string `yaml:"driver"`
	ConnectionString string `yaml:"connectionString"`
	IP               string `yaml:"IP"`
	Port             string `yaml:"port"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Database         string `yaml:"database"`
}

type Application struct {
	IntervalBetweenRequestInSeconds int            `yaml:"intervalBetweenRequestInSeconds"`
	PertinentTags                   map[int]string `yaml:"pertinentTags"`
	LogFilepath                     string         `yaml:"logFilepath"`
	DatesPersistenceFilepath        string         `yaml:"datesPersistenceFilepath"`
	NumberParallelTags              int            `yaml:"numberParallelTags"`
	DataRecoveryPeriodInHours       int            `yaml:"dataRecoveryPeriodInHours"`
	Context                         string         `yaml:"context"`
}

type Query struct {
	Mapping map[int]string `yaml:"mapping"`
}

type IntegrationKNoTConfig struct {
	UserToken               string `yaml:"user_token"`
	URL                     string `yaml:"url"`
	EventRoutingKeyTemplate string `yaml:"event_routing_key_template"`
}
