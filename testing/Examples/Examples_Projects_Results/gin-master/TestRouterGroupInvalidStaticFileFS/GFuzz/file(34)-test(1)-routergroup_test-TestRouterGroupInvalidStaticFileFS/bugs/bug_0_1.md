# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRouterGroupInvalidStaticFileFS
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/routergroup_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/routergroup.go:183
```go
172 ...
173 
174 // Gin by default uses: gin.Dir()
175 func (group *RouterGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem) IRoutes {
176 	return group.staticFileHandler(relativePath, func(c *Context) {
177 		c.FileFromFS(filepath, fs)
178 	})
179 }
180 
181 func (group *RouterGroup) staticFileHandler(relativePath string, handler HandlerFunc) IRoutes {
182 	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
183 		panic("URL parameters can not be used when serving a static file")           // <-------
184 	}
185 	group.GET(relativePath, handler)
186 	group.HEAD(relativePath, handler)
187 	return group.returnObj()
188 }
189 
190 // Static serves files from the given file system root.
191 // Internally a http.FileServer is used, therefore http.NotFound is used instead
192 // of the Router's NotFound handler.
193 // To use the operating system's file system implementation,
194 
195 ...
```


## Replay
**Replaying was not run**.

