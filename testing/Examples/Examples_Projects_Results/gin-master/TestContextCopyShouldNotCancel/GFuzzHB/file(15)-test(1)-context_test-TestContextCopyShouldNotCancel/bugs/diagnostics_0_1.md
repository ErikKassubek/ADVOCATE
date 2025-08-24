# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextCopyShouldNotCancel
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go:3084
```go
3073 ...
3074 
3075 	}()
3076 
3077 	addr := strings.Split(l.Addr().String(), ":")
3078 	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/", addr[len(addr)-1]))
3079 	if err != nil {
3080 		t.Error(fmt.Errorf("request error: %w", err))
3081 		return
3082 	}
3083 
3084 	close(ensureRequestIsOver)           // <-------
3085 
3086 	if res.StatusCode != http.StatusOK {
3087 		t.Error(fmt.Errorf("unexpected status code: %s", res.Status))
3088 		return
3089 	}
3090 
3091 	wg.Wait()
3092 }
3093 
3094 func TestContextAddParam(t *testing.T) {
3095 
3096 ...
```


###  Channel: Receive
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go:3050
```go
3039 ...
3040 
3041 	r.GET("/", func(ginctx *Context) {
3042 		wg.Add(1)
3043 
3044 		ginctx = ginctx.Copy()
3045 
3046 		// start async goroutine for calling srv
3047 		go func() {
3048 			defer wg.Done()
3049 
3050 			<-ensureRequestIsOver // ensure request is done           // <-------
3051 
3052 			req, err := http.NewRequestWithContext(ginctx, http.MethodGet, srv.URL, nil)
3053 			must(err)
3054 
3055 			res, err := http.DefaultClient.Do(req)
3056 			if err != nil {
3057 				t.Error(fmt.Errorf("request error: %w", err))
3058 				return
3059 			}
3060 
3061 
3062 ...
```


## Replay
**Replaying was not run**.

