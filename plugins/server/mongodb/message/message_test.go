package message

import (
	"encoding/hex"
	"testing"
)

func TestIsFlagSet(t *testing.T) {

	b, _ := hex.DecodeString("38000000")
	if isFlagSetUInt32(b, 0, 33) {
		t.Fatalf("isFlagSetUInt32 should be false, because flagpos larger than 31")
	}
	if isFlagSetUInt32(b, 0, -1) {
		t.Fatalf("isFlagSetUInt32 should be false, because flagpos smaller than 0")
	}

	// matching byte: 00010000. flag pos 4
	b, _ = hex.DecodeString("38000000")
	if !isFlagSetInt32(b, 0, 4) {
		t.Fatalf("isFlagSetInt32 should return true")
	}

	// matching byte 	00000000 00000000 00010000 00000000. flag pos 20
	// b 38001111		00111000 00000000 00010001 00010001
	// 									     ^
	b, _ = hex.DecodeString("38001111")
	if !isFlagSetUInt32(b, 0, 20) {
		t.Fatalf("isFlagSetUInt32 should return true")
	}

}
