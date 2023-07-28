package entities

type CapturedData struct {
	ID   string `json:"id"`
	Data []Data `json:"data"`
}

type Data struct {
	SensorID  int     `json:"sensorId"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}
