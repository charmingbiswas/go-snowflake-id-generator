package snowflake

import "time"

type parsed struct {
	TimeStamp time.Time
	MachineId int64
	Sequence  int64
}

func Parse(snowflakeId int64) *parsed {
	timeStamp := (snowflakeId >> timestampShift) + epoch
	return &parsed{
		TimeStamp: time.UnixMilli(timeStamp),
		MachineId: (snowflakeId >> nodeShift) & maxNodes,
		Sequence:  snowflakeId & maxSequence,
	}
}
