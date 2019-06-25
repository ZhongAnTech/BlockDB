package message

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewMsgMessage(t *testing.T) {

	//data1 := "82000000d9acb76000000000dd07000000000000006d0000001069736d61737465720001000000027361736c537570706f727465644d65636873000b00000061646d696e2e726f6f740002246462000600000061646d696e00032472656164507265666572656e63650017000000026d6f646500080000007072696d617279000000"
	//if b := msgTester(t, data1); b != nil {
	//	fmt.Println(string(b))
	//}

	data2 := "f8000000fe09e55600000000dd07000000000000007a00000002696e73657274000600000070726f787900086f7264657265640001036c736964001e000000056964001000000004f565892c2495440ab9dc0ea84c98f2b100022464620004000000756e6900032472656164507265666572656e63650017000000026d6f646500080000007072696d6172790000000168000000646f63756d656e7473005a000000075f6964005d11dee343096d19ac7c070c02696e736572745f74696d650014000000323031392d30362d32352031363a34343a313900026461746100010000000002647269766572000800000070796d6f6e676f0000"
	if b := msgTester(t, data2); b != nil {
		fmt.Println(string(b))
	}

}

func msgTester(t *testing.T, dataHex string) []byte {

	dataBytes, _ := hex.DecodeString(dataHex)

	header, err := DecodeHeader(dataBytes)
	if err != nil {
		t.Fatalf("decode header error: %v", err)
		return nil
	}
	msg, err := NewMsgMessage(header, dataBytes)
	if err != nil {
		t.Fatalf("create msg message error: %v", err)
		return nil
	}

	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json marshal message error: %v", err)
		return nil
	}

	return b
}
