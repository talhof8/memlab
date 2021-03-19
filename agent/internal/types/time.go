package types

import (
	"gopkg.in/guregu/null.v3"
	"time"
)

func JsonTimeFromMillisecondTimestamp(timestamp int64) null.Time {
	return JsonTimeFromTimestamp(timestamp / 1000)
}

func JsonTimeFromTimestamp(timestamp int64) null.Time {
	return null.TimeFrom(time.Unix(timestamp, 0).UTC())
}
