# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextError
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context.go:246
```go
235 ...
236 
237 /************************************/
238 
239 // Error attaches an error to the current context. The error is pushed to a list of errors.
240 // It's a good idea to call Error for each error that occurred during the resolution of a request.
241 // A middleware can be used to collect all the errors and push them to a database together,
242 // print a log, or append it in the HTTP response.
243 // Error will panic if err is nil.
244 func (c *Context) Error(err error) *Error {
245 	if err == nil {
246 		panic("err is nil")           // <-------
247 	}
248 
249 	var parsedError *Error
250 	ok := errors.As(err, &parsedError)
251 	if !ok {
252 		parsedError = &Error{
253 			Err:  err,
254 			Type: ErrorTypePrivate,
255 		}
256 	}
257 
258 ...
```


## Replay
**Replaying was not run**.

