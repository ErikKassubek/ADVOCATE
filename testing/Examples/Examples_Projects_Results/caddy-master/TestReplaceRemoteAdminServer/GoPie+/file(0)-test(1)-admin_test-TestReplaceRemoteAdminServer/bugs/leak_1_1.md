# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestReplaceRemoteAdminServer
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/admin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/caddy-master/logging.go:773
```go
762 ...
763 
764 	default:
765 		return nil, fmt.Errorf("unrecognized log level: %s", level)
766 	}
767 }
768 
769 // Log returns the current default logger.
770 func Log() *zap.Logger {
771 	defaultLoggerMu.RLock()
772 	defer defaultLoggerMu.RUnlock()
773 	return defaultLogger.logger           // <-------
774 }
775 
776 var (
777 	coloringEnabled  = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "xterm-mono"
778 	defaultLogger, _ = newDefaultProductionLog()
779 	defaultLoggerMu  sync.RWMutex
780 )
781 
782 var writers = NewUsagePool()
783 
784 
785 ...
```


## Replay
**Replaying was not run**.

