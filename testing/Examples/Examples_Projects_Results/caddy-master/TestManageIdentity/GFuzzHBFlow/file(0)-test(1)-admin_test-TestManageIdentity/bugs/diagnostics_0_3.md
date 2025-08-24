# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestManageIdentity
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/github.com/caddyserver/certmagic@v0.23.0/cache.go:144
```go
133 ...
134 
135 	certCache.options = opts
136 	certCache.optionsMu.Unlock()
137 }
138 
139 // Stop stops the maintenance goroutine for
140 // certificates in certCache. It blocks until
141 // stopping is complete. Once a cache is
142 // stopped, it cannot be reused.
143 func (certCache *Cache) Stop() {
144 	close(certCache.stopChan) // signal to stop           // <-------
145 	<-certCache.doneChan      // wait for stop to complete
146 }
147 
148 // CacheOptions is used to configure certificate caches.
149 // Once a cache has been created with certain options,
150 // those settings cannot be changed.
151 type CacheOptions struct {
152 	// REQUIRED. A function that returns a configuration
153 	// used for managing a certificate, or for accessing
154 	// that certificate's asset storage (e.g. for
155 
156 ...
```


###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/github.com/caddyserver/certmagic@v0.23.0/maintain.go:69
```go
58 ...
59 
60 	ocspTicker := time.NewTicker(certCache.options.OCSPCheckInterval)
61 	certCache.optionsMu.RUnlock()
62 
63 	log.Info("started background certificate maintenance")
64 
65 	ctx, cancel := context.WithCancel(context.Background())
66 	defer cancel()
67 
68 	for {
69 		select {           // <-------
70 		case <-renewalTicker.C:
71 			err := certCache.RenewManagedCertificates(ctx)
72 			if err != nil {
73 				log.Error("renewing managed certificates", zap.Error(err))
74 			}
75 		case <-ocspTicker.C:
76 			certCache.updateOCSPStaples(ctx)
77 		case <-certCache.stopChan:
78 			renewalTicker.Stop()
79 			ocspTicker.Stop()
80 
81 ...
```


## Replay
**Replaying was not run**.

