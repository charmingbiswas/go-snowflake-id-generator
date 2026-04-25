package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	EPOCH           = int64(1700000000000)
	MACHINE_ID_BITS = 10
	SEQUENCE_BITS   = 12
	MAX_NODES       = int64(-1) ^ (int64(-1) << MACHINE_ID_BITS)
	MAX_SEQUENCE    = int64(-1) ^ (int64(-1) << SEQUENCE_BITS)
	NODE_SHIFT      = SEQUENCE_BITS
	TIMESTAMP_SHIFT = MACHINE_ID_BITS + SEQUENCE_BITS
)

/*
	int64(-1) ^ (int64(-1) << N) is a classic bitmask trick. -1 is all 1-bits.
	Shifting left by N clears the lower N bits, then XOR-ing with -1 flips everything — leaving exactly N bits set.
	This gives you the maximum value for that field without hardcoding.
*/

type monotonicClock struct {
	startWall int64     // wall clock time
	startMono time.Time // monotonic clock time
}

type node struct {
	mu            sync.Mutex
	machineId     int64
	sequence      int64
	lastTimeStamp int64
	monoClock     *monotonicClock
}

func NewNode(machineId int64) (*node, error) {
	if machineId < 0 || machineId > MAX_NODES {
		return nil, fmt.Errorf("machine Id must be between 0 and %d", MAX_NODES)
	}
	return &node{machineId: machineId, monoClock: newMonoClock()}, nil
}

func (n *node) GenerateSnowflakeId() (int64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := n.monoClock.nowMs() // get's the monotonic clock value, never goes back hence not affected by NTP syncs

	if now < n.lastTimeStamp {
		// clock drifted backwards
		// this case should ideally not occur since we are using monotonic clocks
		// keeping it in as a safety fallback
		return 0, fmt.Errorf("clock moved backwards by %d, please try again", n.lastTimeStamp-now)
	}

	if now == n.lastTimeStamp {
		n.sequence = (n.sequence + 1) & MAX_SEQUENCE
		if n.sequence == 0 {
			// meaning you ran out of sequences for the current timestamp
			// wait until next millisecond
			for now <= n.lastTimeStamp {
				now = n.monoClock.nowMs()
			}
		}
	} else {
		n.sequence = 0 // Reset the sequence
	}

	n.lastTimeStamp = now

	/*
		(now - epoch) << 22 — the timestamp occupies bits 63–22
		nodeID << 12 — the machine ID occupies bits 21–12
		sequence — sits in the lowest 12 bits (no shift needed)
		The bitwise OR (|) merges all three into a single int64.
	*/
	snowflakeId := now<<TIMESTAMP_SHIFT | n.machineId<<NODE_SHIFT | n.sequence

	return snowflakeId, nil
}

func GetMachineIdFromEnv() (int64, error) {
	value := os.Getenv(snowflakeMachineIDEnvKey)
	if value == "" {
		return 0, fmt.Errorf("no env variable with key %s found", snowflakeMachineIDEnvKey)
	}

	splitValues := strings.Split(value, "-") // env variable will be set by kubernetes statefulset which looks like pod-abcd-0

	id, err := strconv.ParseInt(splitValues[len(splitValues)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s key found: %d", snowflakeMachineIDEnvKey, id)
	}

	return id, nil
}

func newMonoClock() *monotonicClock {
	now := time.Now()
	return &monotonicClock{
		startWall: now.UnixMilli() - EPOCH,
		startMono: now,
	}
}

// returns milliseconds passed using monotonic component
// this is immune to clokc skew and clock drifts
func (mc *monotonicClock) nowMs() int64 {
	// time.Since used the monotonic clock
	elapsedTime := time.Since(mc.startMono).Milliseconds()
	return mc.startWall + elapsedTime // we add the elaspsedTime to startWall since monotonic duration by itself has no value. To generate a snowflake ID it need real clock time.
}
