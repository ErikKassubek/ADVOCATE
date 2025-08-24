# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestLoadConcurrent
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin.go:454
```go
443 ...
444 
445 		IdleTimeout:       60 * time.Second,
446 		MaxHeaderBytes:    1024 * 64,
447 	}
448 	serverMu.Unlock()
449 
450 	adminLogger := Log().Named("admin")
451 	go func() {
452 		serverMu.Lock()
453 		server := localAdminServer
454 		serverMu.Unlock()           // <-------
455 		if err := server.Serve(ln.(net.Listener)); !errors.Is(err, http.ErrServerClosed) {
456 			adminLogger.Error("admin server shutdown for unknown reason", zap.Error(err))
457 		}
458 	}()
459 
460 	adminLogger.Info("admin endpoint started",
461 		zap.String("address", addr.String()),
462 		zap.Bool("enforce_origin", cfg.Admin.EnforceOrigin),
463 		zap.Array("origins", loggableURLArray(handler.allowedOrigins)))
464 
465 
466 ...
```


## Replay
**Replaying was not run**.

