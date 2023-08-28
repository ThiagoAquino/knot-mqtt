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

type Query struct {
	Mapping map[int]string `yaml:"mapping"`
}

type IntegrationKNoTConfig struct {
	UserToken               string `yaml:"user_token"`
	URL                     string `yaml:"url"`
	EventRoutingKeyTemplate string `yaml:"event_routing_key_template"`
}

type MqttConfig struct {
	MqttBroker   string `yaml:"mqttBroker"`
	MqttClientID string `yaml:"mqttClientID"`
	Topic        string `yaml:"topic"`
	LogFilepath  string `yaml:"logFilepath"`
	AmountTags   int    `yaml:"amountTags"`
	MqttQoS      byte   `yaml:"mqttQoS"`
	MqttUser     string `yaml:"mqttUser"`
	MqttPass     string `yaml:"mqttPass"`
}
