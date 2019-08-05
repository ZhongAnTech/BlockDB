package ogws

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeZone(t *testing.T) {
	l := int64(1564453129000)
	strt := time.Unix(0, l*int64(1000000))
	loc, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	strt = strt.In(loc)
	fmt.Println(strt.Format("2006-01-02 15:04:05"))
}
