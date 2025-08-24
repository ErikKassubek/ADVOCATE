# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRouterGroupInvalidStatic
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/routergroup_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/routergroup.go:205
```go
194 ...
195 
196 //	router.Static("/static", "/var/www")
197 func (group *RouterGroup) Static(relativePath, root string) IRoutes {
198 	return group.StaticFS(relativePath, Dir(root, false))
199 }
200 
201 // StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
202 // Gin by default uses: gin.Dir()
203 func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
204 	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
205 		panic("URL parameters can not be used when serving a static folder")           // <-------
206 	}
207 	handler := group.createStaticHandler(relativePath, fs)
208 	urlPattern := path.Join(relativePath, "/*filepath")
209 
210 	// Register GET and HEAD handlers
211 	group.GET(urlPattern, handler)
212 	group.HEAD(urlPattern, handler)
213 	return group.returnObj()
214 }
215 
216 
217 ...
```


## Replay
**Replaying was not run**.

