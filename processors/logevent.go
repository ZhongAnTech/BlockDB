package processors

type LogEvent struct {
	Identity  string `json:"identity"`
	Ip        string `json:"ip"`
	Timestamp uint32 `json:"timestamp"`
	Data      string `json:"data"`
	Before    string `json:"before"`
	After     string `json:"after"`
}
