package processors

type LogEvent struct {
	Identity  string      `json:"identity"`
	Type      string      `json:"type"`
	Ip        string      `json:"ip"`
	Timestamp int         `json:"timestamp"`
	Data      interface{} `json:"data"`
	Before    string      `json:"before"`
	After     string      `json:"after"`
}
