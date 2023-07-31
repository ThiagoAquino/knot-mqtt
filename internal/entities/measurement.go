package entities

type Row struct {
	ID        int     `json:"sensorId"`
	Value     float64 `json:"value"`
	Timestamp string  `json:"timestamp"`
}

type Statement struct {
	Table     string `json:"table"`
	Timestamp string `json:"timestamp"`
}

type CapturedData struct {
	ID   int   `json:"id"`
	Rows []Row `json:"data"`
}
