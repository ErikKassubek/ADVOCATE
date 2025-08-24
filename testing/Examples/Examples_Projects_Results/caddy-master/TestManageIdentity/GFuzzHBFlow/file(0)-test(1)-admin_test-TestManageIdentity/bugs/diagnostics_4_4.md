# Diagnostics: Actual Receive on Closed Channel

During the execution of the program, a receive on a closed channel occurred.


## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestManageIdentity
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found diagnostics are located at the following positions:

###  Channel: Close
-> /home/ilian/Projects/go/pkg/mod/github.com/caddyserver/certmagic@v0.23.0/maintain.go:81
```go
70 ...
71 
72 			if err != nil {
73 				log.Error("renewing managed certificates", zap.Error(err))
74 			}
75 		case <-ocspTicker.C:
76 			certCache.updateOCSPStaples(ctx)
77 		case <-certCache.stopChan:
78 			renewalTicker.Stop()
79 			ocspTicker.Stop()
80 			log.Info("stopped background certificate maintenance")
81 			close(certCache.doneChan)           // <-------
82 			return
83 		}
84 	}
85 }
86 
87 // RenewManagedCertificates renews managed certificates,
88 // including ones loaded on-demand. Note that this is done
89 // automatically on a regular basis; normally you will not
90 // need to call this. This method assumes non-interactive
91 // mode (i.e. operating in the background).
92 
93 ...
```


###  Channel: Receive
-> /home/ilian/Projects/go/pkg/mod/github.com/caddyserver/certmagic@v0.23.0/cache.go:145
```go
134 ...
135 
136 	certCache.optionsMu.Unlock()
137 }
138 
139 // Stop stops the maintenance goroutine for
140 // certificates in certCache. It blocks until
141 // stopping is complete. Once a cache is
142 // stopped, it cannot be reused.
143 func (certCache *Cache) Stop() {
144 	close(certCache.stopChan) // signal to stop
145 	<-certCache.doneChan      // wait for stop to complete           // <-------
146 }
147 
148 // CacheOptions is used to configure certificate caches.
149 // Once a cache has been created with certain options,
150 // those settings cannot be changed.
151 type CacheOptions struct {
152 	// REQUIRED. A function that returns a configuration
153 	// used for managing a certificate, or for accessing
154 	// that certificate's asset storage (e.g. for
155 	// OCSP staples, etc). The returned Config MUST
156 
157 ...
```


## Replay
**Replaying was not run**.

