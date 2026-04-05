package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	epoch          = int64(1700000000000)
	machineIdBits  = 10
	sequenceBits   = 12
	maxNodes       = int64(-1) ^ (int64(-1) << machineIdBits)
	maxSequence    = int64(-1) ^ (int64(-1) << sequenceBits)
	nodeShift      = sequenceBits
	timestampShift = machineIdBits + sequenceBits
)

/*
	int64(-1) ^ (int64(-1) << N) is a classic bitmask trick. -1 in two's complement is all 1-bits.
	Shifting left by N clears the lower N bits, then XOR-ing with -1 flips everything — leaving exactly N bits set.
	This gives you the maximum value for that field without hardcoding magic numbers.
*/

type node struct {
	mu            sync.Mutex
	machineId     int64
	sequence      int64
	lastTimeStamp int64
}

func NewNode(machineId int64) (*node, error) {
	if machineId < 0 || machineId > maxNodes {
		return nil, fmt.Errorf("machine Id must be between 0 and %d", maxNodes)
	}
	return &node{machineId: machineId}, nil
}

func (n *node) GenerateSnowflakeId() (int64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < n.lastTimeStamp {
		// clock drifted backwards
		return 0, fmt.Errorf("clock moved backwards by %d, please try again", n.lastTimeStamp-now)
	}

	if now == n.lastTimeStamp {
		n.sequence = (n.sequence + 1) & maxSequence
		if n.sequence == 0 {
			// meaning you ran out of sequences for the current timestamp
			// wait until next millisecond
			for now <= n.lastTimeStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.sequence = 0
	}

	n.lastTimeStamp = now

	/*
		(now - epoch) << 22 — the timestamp occupies bits 63–22
		nodeID << 12 — the machine ID occupies bits 21–12
		sequence — sits in the lowest 12 bits (no shift needed)
		The bitwise OR (|) merges all three into a single int64.
	*/
	snowflakeId := (now-epoch)<<timestampShift | n.machineId<<nodeShift | n.sequence

	return snowflakeId, nil
}

func GetMachineIdFromEnv() (int64, error) {
	value := os.Getenv(snowflakeMachineIDEnvKey)
	if value == "" {
		return 0, fmt.Errorf("no env variable with key %s found", snowflakeMachineIDEnvKey)
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s key found: %d", snowflakeMachineIDEnvKey, id)
	}

	return id, nil
}
