package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestEmitNonConcurrent(t *testing.T) {
	pipeline := New(WithDefault())
	pipeline.Next("test", func(i interface{}) (interface{}, error) {
		return i, nil
	}, 1, 1024)

	arr := []string{"test1", "test2", "test3", "test4"}

	counter := 0

	// a non-concurrent pipeline will output the values
	// to the sink in the same order they were emitted on the source
	sink := pipeline.Emit("test1", "test2", "test3", "test4")

	for actual := range sink {
		expect := arr[counter]
		if actual != expect {
			t.Errorf("expected emitted value to be %+v but was %+v", expect, actual)
		}
		counter++
	}
}

func TestEmitManyStages(t *testing.T) {
	pipeline := New(WithDefault())

	pipeline.Next("stage1", func(i interface{}) (interface{}, error) {
		return strings.ToUpper(i.(string)), nil
	}, 2, 1024)
	pipeline.Next("stage1", func(i interface{}) (interface{}, error) {
		return fmt.Sprintf("--%s--", i.(string)), nil
	}, 1, 1024)
	pipeline.Next("stage1", func(i interface{}) (interface{}, error) {
		return fmt.Sprintf("123%s456", i.(string)), nil
	}, 3, 1024)

	expect := "123--HELLO--456"

	sink := pipeline.Emit("hello", "hello", "hello")

	for actual := range sink {
		if actual != expect {
			t.Errorf("expected emitted value to be %s but was %+v", expect, actual)
		}
	}
}

func TestDumpsErrors(t *testing.T) {

	pipeline := New(WithDefault())
	pipeline.OnError(func(e error) {
		// do nothing
	})

	pipeline.Next("test", func(i interface{}) (interface{}, error) {
		if i == nil {
			return nil, errors.New("test")
		}
		return i, nil
	}, 1, 1024)

	emitted := 0
	sink := pipeline.Emit("one", "two", nil)
	for range sink {
		emitted++
	}

	if emitted != 2 {
		t.Errorf("expected pipeline to emit two values but emitted %d values", emitted)
	}
}
