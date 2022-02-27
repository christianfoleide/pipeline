package pipeline

import (
	"strings"
	"testing"
)

func TestEmitFromChannel(t *testing.T) {
	source := make(chan interface{})
	sendCount := 5

	go func() {
		defer close(source)
		for i := 0; i < sendCount; i++ {
			source <- "value"
		}
	}()

	p := New(WithDefault())

	p.Next(func(i interface{}) (interface{}, error) {
		data, _ := i.(string)
		return strings.ToUpper(data), nil
	}, 3)

	expectValue := "VALUE"
	emitCount := 0

	for v := range p.EmitFromChannel(source) {
		if v != expectValue {
			t.Errorf("expected value to be %s but was %+v", expectValue, v)
		}
		emitCount++
	}

	if emitCount != sendCount {
		t.Errorf("expected number received values to be %d but was %d", sendCount, emitCount)
	}
}
