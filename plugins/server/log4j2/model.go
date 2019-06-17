package log4j2

type Log4j2SocketEventInstant struct {
	timestamp uint32 `json:"epochSecond"`
}

type Log4j2SocketEvent struct {
	loggerName string                   `json:"loggerName"`
	message    string                   `json:"message"`
	instant    Log4j2SocketEventInstant `json:"instant"`
	contextMap interface{}              `json:"contextMap"`
}
