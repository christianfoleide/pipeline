package pipeline

import (
	"log"
	"sync"
	"time"
)

// stageFunc is the function that is called on a value emitted in
// a single stage of the pipeline.
type stageFunc func(interface{}) (interface{}, error)

// if a stage performs a stageFunc that returns an error,
// that error is passed to this function.
type errFunc func(error)

func defaultErrFunc(err error) {
	log.Printf("stageFunc returned a non-nil error: %+v", err)
}

// cancelFunc is the function returned by either Emit or
// EmitWithDelay, making the client able to cancel
// the pipeline.
type cancelFunc func()

// Stage represents a single processing stage in the pipeline.
type stage struct {
	workerCount   int
	cancelOnError bool
	fn            stageFunc
	errFn         errFunc
	in            chan interface{}
	out           chan interface{}
}

// Run starts the processing stage, creating workerCount goroutines.
// Each goroutine receives on this stage's inbound channel, performs
// the stageFunc on the received values, and emits them on the stage's
// outbound channel.
func (s *stage) Run() {

	wg := &sync.WaitGroup{}
	wg.Add(s.workerCount)
	go func() {
		wg.Wait()
		close(s.out)
	}()

	for i := 0; i < s.workerCount; i++ {
		go func() {
			defer wg.Done()
			for v := range s.in {
				res, err := s.fn(v)
				if err != nil {
					s.errFn(err)
					continue
				}
				s.out <- res
			}
		}()
	}
}

type Pipeline struct {
	source     chan interface{}
	sink       chan interface{}
	cancelChan chan struct{}

	errFn   errFunc
	stages  []*stage
	bufSize int
}

type PipelineOpts struct {
	ChanBufSize int
}

// WithDefault sets the channel bufSize of the pipeline to 1024.
func WithDefault() *PipelineOpts {
	return &PipelineOpts{
		ChanBufSize: 1024,
	}
}

// New returns a new Pipeline.
func New(opts *PipelineOpts) *Pipeline {
	return &Pipeline{
		source:  make(chan interface{}, opts.ChanBufSize),
		sink:    make(chan interface{}, opts.ChanBufSize),
		stages:  make([]*stage, 0),
		errFn:   defaultErrFunc,
		bufSize: opts.ChanBufSize,
	}
}

// OnError sets the callback function to be called on
// any error values returned by a stageFunc on any given stage.
// Error values will not emit to the downstream stage, but
// instead stop at the point they were produced.
func (p *Pipeline) OnError(fn errFunc) {
	p.errFn = fn
}

// Next adds a processing stage to the pipeline.
// The stage spawns routineCount goroutines, and performs fn on
// each piece of data it receives from the upstream stage, if any
// such stage exists.
// The data flows through the pipeline in the order Next is called.
func (p *Pipeline) Next(fn stageFunc, routineCount int) {
	if len(p.stages) == 0 {
		stage := &stage{
			workerCount: routineCount,
			fn:          fn,
			in:          p.source,
			out:         p.sink,
		}
		p.stages = append(p.stages, stage)
		stage.Run()
		return

	}
	join := make(chan interface{}, p.bufSize)
	last := p.stages[len(p.stages)-1]
	last.out = join // the last stage is no longer emitting to the pipeline sink.

	stage := &stage{
		workerCount: routineCount,
		fn:          fn,
		in:          join,   // this new stage will receive from the upstream stage.
		out:         p.sink, // instead, this new stage will emit to the sink.
	}
	p.stages = append(p.stages, stage)
	stage.Run()
}

// Emit emits the given values on the pipeline's source channel.
// It returns a channel representing the pipeline's sink channel,
// where resulting values from the entire pipeline chain is emitted.
func (p *Pipeline) Emit(values ...interface{}) <-chan interface{} {

	p.mustHaveStages()

	go func() {
		defer close(p.source)
		for _, v := range values {
			p.source <- v
		}
	}()
	return p.sink
}

// EmitWithDelay emits the values on the pipeline's source channel
// in intervals of the passed duration.
func (p *Pipeline) EmitWithDelay(dur time.Duration, values ...interface{}) <-chan interface{} {

	p.mustHaveStages()

	go func() {
		defer close(p.source)
		for _, v := range values {
			select {
			case <-time.After(dur):
				p.source <- v
			}
		}
	}()
	return p.sink
}

// Ensure that when a client makes a call to Emit or EmitWithDelay,
// the pipeline contains processing stages, and set the error callback
// on each of them in case OnError was called before calls to pipeline.Next(...).
func (p *Pipeline) mustHaveStages() {

	if len(p.stages) == 0 {
		log.Fatalln("no processing stages in pipeline")
	}

	for _, stage := range p.stages {
		stage.errFn = p.errFn
	}
}
