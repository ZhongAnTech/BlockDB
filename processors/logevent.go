package processors

type LogEvent struct {
	Identity  string `json:"identity"`
	Ip        string `json:"ip"`
	Timestamp int    `json:"timestamp"`
	Data      string `json:"data"`
	Before    string `json:"before"`
	After     string `json:"after"`
}
