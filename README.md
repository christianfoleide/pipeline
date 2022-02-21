## Minimal concurrent data pipeline

```
$ go get -u github.com/christianfoleide/pipeline
```

### Example

```go
func main() {

   p := pipeline.New()

   p.Next("stage1", func(data interface{}) (interface{}, error) {

      // data is from upstream stage
      // returned data is sent to a downstream stage

   }, 3) // 3 goroutines

   p.Next("stage2", func(data interface{}) (interface{}, error) {

      // ----..----

   }, 2) // 2 goroutines

   p.OnError(func(err error) {
      // handle errors on any stage
   })

   for v := range p.Emit(/* source data */) {

      // v is the resulting values
      // altered by the entire pipeline chain

   }
}

```

### Limitations

- No cancellation
   - And hence no draining of remaining data
- Errors are handled one way