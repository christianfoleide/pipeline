## Minimal concurrent data pipeline

```
$ go get -u github.com/christianfoleide/pipeline
```

Basically a glorified linked-list of channels


### Example

```go
func main() {

   p := pipeline.New()

   // adding a stage

   p.Next(func(data interface{}) (interface{}, error) {

      // data is from upstream stage
      // returned data is sent to a downstream stage

   }, 3) // 3 goroutines

   p.Next(func(data interface{}) (interface{}, error) {

      // do something with data

   }, 2) // 2 goroutines

   p.OnError(func(err error) {
      // handle errors on any stage
   })

   // emit data

   for v := range p.Emit(/* source data */) {

      // v is a single value modified by the entire
      // pipeline chain

   }
}

```

### Limitations
- No options to stop on errors:
   - The error is passed to the OnError handler, and will not go past the stage it was produced.