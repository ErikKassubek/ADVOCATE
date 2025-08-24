# Leak: Leak on select without possible partner

The analyzer detected a Leak on a select without a possible partner.
A Leak on a select is a situation, where a select is still blocking at the end of the program.
The analyzer could not find a partner for the stuck operation, which would resolve the leak.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestReplaceRemoteAdminServer
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Select:
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

