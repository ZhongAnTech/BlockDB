package processors

type LogEvent struct {
	Identity   string      `json:"identity"`
	Type       string      `json:"type"`
	Ip         string      `json:"ip"`
	PrimaryKey string      `json:"primary_key"`
	Timestamp  int64       `json:"timestamp"`
	Data       interface{} `json:"data"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
}
