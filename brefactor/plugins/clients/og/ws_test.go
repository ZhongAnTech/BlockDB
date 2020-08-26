package og

import (
	"testing"
	"time"
)

func TestWs(t *testing.T) {
	for {

		time.Sleep(3 * time.Second)
		height := Http()
		hashByHeight(height)

	}
}
