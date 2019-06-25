package log4j2

type Log4j2SocketEventInstant struct {
	Timestamp int `json:"epochSecond"`
}

type Log4j2SocketEvent struct {
	Level      string                   `json:"level"`
	LoggerName string                   `json:"loggerName"`
	Message    string                   `json:"message"`
	Instant    Log4j2SocketEventInstant `json:"instant"`
	ContextMap map[string]interface{}   `json:"contextMap"`
}
