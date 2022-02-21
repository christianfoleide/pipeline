package pipeline

import (
	"testing"
	"time"
)

func TestDelayEmit(t *testing.T) {

	pipeline := New()
	pipeline.Next("test", func(i interface{}) (interface{}, error) {
		return i, nil
	}, 1)

	start := time.Now()
	sink := pipeline.EmitWithDelay(time.Millisecond*250, "", "", "", "")
	for range sink {
		// do nothing
	}

	elapsed := time.Now().Sub(start)

	if !(elapsed.Milliseconds() > 950) && !(elapsed.Milliseconds() < 1050) {
		t.Errorf("expected elapsed time to be between 900-1100 ms but was %d", elapsed.Milliseconds())
	}
}
