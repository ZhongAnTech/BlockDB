package og

import (
	"time"

	"github.com/sirupsen/logrus"
)

const checkInterval = time.Second*5

func (o *OgProcessor) fetchData() {

	pingTicker := time.NewTicker(checkInterval)
	defer pingTicker.Stop()
outside:
	for {
		select {
		case <-pingTicker.C:
			limit:= int64(100)
			offset:= int64(0)
			count :=int64(10)
			for offset<count {
				expiredData, total, err := o.originalDataProcessor.GetExpired(time.Second*10, limit, offset)
				count = total
				if err != nil {
					logrus.WithError(err).Error("read data err")
				}
				for _, v := range expiredData {
					err := o.EnqueueSendToLedger(v.Data)
					if err != nil {
						logrus.WithField("data", v).WithError(err).Warn("send data err")
					} else {
						o.originalDataProcessor.DeleteOne(v.Hash)
					}
				}
				offset += limit
			}


		case <-o.quitFetch:
			break outside
		}
	}
	logrus.Info("OgProcessor fetchdata  stopped")
}