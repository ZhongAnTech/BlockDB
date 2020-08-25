package instruction

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func toDoc(v interface{}) (doc bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func ts() int64 {
	return time.Now().UnixNano() / 1_000_000
}
