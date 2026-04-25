package snowflake

import "time"

type parsed struct {
	TimeStamp time.Time
	MachineId int64
	Sequence  int64
}

func Parse(snowflakeId int64) *parsed {
	timeStamp := (snowflakeId >> TIMESTAMP_SHIFT) + EPOCH
	return &parsed{
		TimeStamp: time.UnixMilli(timeStamp),
		MachineId: (snowflakeId >> NODE_SHIFT) & MAX_NODES,
		Sequence:  snowflakeId & MAX_SEQUENCE,
	}
}
