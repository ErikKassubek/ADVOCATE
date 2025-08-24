# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestContextRenderRedirectWithRelativePath
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/context_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/render/redirect.go:22
```go
11 ...
12 
13 type Redirect struct {
14 	Code     int
15 	Request  *http.Request
16 	Location string
17 }
18 
19 // Render (Redirect) redirects the http request to new location and writes redirect response.
20 func (r Redirect) Render(w http.ResponseWriter) error {
21 	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
22 		panic(fmt.Sprintf("Cannot redirect with status code %d", r.Code))           // <-------
23 	}
24 	http.Redirect(w, r.Request, r.Location, r.Code)
25 	return nil
26 }
27 
28 // WriteContentType (Redirect) don't write any ContentType.
29 func (r Redirect) WriteContentType(http.ResponseWriter) {}
30 
```


## Replay
**Replaying was not run**.

