package types

import "time"

func TimeFromMillisecondTimestamp(timestamp int64) time.Time {
	return TimeFromTimestamp(timestamp / 1000)
}

func TimeFromTimestamp(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).UTC()
}
