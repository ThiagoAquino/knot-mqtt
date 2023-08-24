package entities

type Row struct {
	Value     interface{} `json:"value"`
	Timestamp string      `json:"timestamp"`
}

type Statement struct {
	Table     string `json:"table"`
	Timestamp string `json:"timestamp"`
}

type CapturedData struct {
	ID   int   `json:"sensorId"`
	Rows []Row `json:"data"`
}
