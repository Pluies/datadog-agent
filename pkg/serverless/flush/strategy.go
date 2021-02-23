package flush

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Strategy is deciding whether the data should be flushed or not at the given moment.
type Strategy interface {
	String() string
	ShouldFlush(moment Moment, t time.Time) bool
}

// Moment represents at which moment we're asking the flush strategy if we
// should flush or not.
// Note that there is no entry for the shutdown of the environment because we always
// flush in this situation.
type Moment string

const (
	// Starting is used to represent the moment the function is starting because
	// it has been invoked.
	Starting Moment = "starting"
	// Stopping is used to represent the moment right after the function has finished
	// its execution.
	Stopping Moment = "stopping"
)

// StrategyFromString returns a flush strategy from the given string.
// Possible values:
//   - start
//   - end
//   - periodically[,milliseconds]
func StrategyFromString(str string) (Strategy, error) {
	switch str {
	case "start":
		return &AtTheStart{}, nil
	case "end":
		return &AtTheEnd{}, nil
	case "periodically":
		return &Periodically{interval: 10 * time.Second}, nil
	}

	if strings.HasPrefix(str, "periodically") && strings.Count(str, ",") == 1 {
		parts := strings.Split(str, ",")
		if msecs, err := strconv.Atoi(parts[1]); err != nil {
			return &AtTheEnd{}, fmt.Errorf("StrategyFromString: can't parse flush strategy: %s", str)
		} else {
			return &Periodically{interval: time.Duration(msecs) * time.Millisecond}, nil
		}
	}

	return &AtTheEnd{}, fmt.Errorf("StrategyFromString: can't parse flush strategy: %s", str)
}

// -----

// AtTheEnd strategy is the simply flushing the data at the end of the execution of the function.
type AtTheEnd struct{}

func (s *AtTheEnd) String() string { return "end" }

// ShouldFlush returns true if this strategy want to flush at the given moment.
func (s *AtTheEnd) ShouldFlush(moment Moment, t time.Time) bool {
	return moment == Stopping
}

// -----

// AtTheStart is the strategy flushing at the start of the execution of the function.
type AtTheStart struct{}

func (s *AtTheStart) String() string { return "start" }

// ShouldFlush returns true if this strategy want to flush at the given moment.
func (s *AtTheStart) ShouldFlush(moment Moment, t time.Time) bool {
	return moment == Starting
}

// -----

// Periodically is the strategy flushing at least every N [nano/micro/milli]seconds
// at the start of the function.
type Periodically struct {
	interval  time.Duration
	lastFlush time.Time
}

// NewPeriodically returns an initialized Periodically flush strategy.
func NewPeriodically(interval time.Duration) *Periodically {
	return &Periodically{interval: interval}
}

func (s *Periodically) String() string {
	return fmt.Sprintf("periodically,%d", s.interval/time.Millisecond)
}

// ShouldFlush returns true if this strategy want to flush at the given moment.
func (s *Periodically) ShouldFlush(moment Moment, t time.Time) bool {
	if moment == Starting {
		now := time.Now()
		if s.lastFlush.Add(s.interval).Before(now) {
			s.lastFlush = now
			return true
		}
	}
	return false
}
