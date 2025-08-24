# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestH2c
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/gin_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/go/pkg/mod/golang.org/x/net@v0.41.0/http2/transport.go:2833
```go
2822 ...
2823 
2824 	}
2825 	cc.setGoAway(f)
2826 	return nil
2827 }
2828 
2829 func (rl *clientConnReadLoop) processSettings(f *SettingsFrame) error {
2830 	cc := rl.cc
2831 	// Locking both mu and wmu here allows frame encoding to read settings with only wmu held.
2832 	// Acquiring wmu when f.IsAck() is unnecessary, but convenient and mostly harmless.
2833 	cc.wmu.Lock()           // <-------
2834 	defer cc.wmu.Unlock()
2835 
2836 	if err := rl.processSettingsNoWrite(f); err != nil {
2837 		return err
2838 	}
2839 	if !f.IsAck() {
2840 		cc.fr.WriteSettingsAck()
2841 		cc.bw.Flush()
2842 	}
2843 	return nil
2844 
2845 ...
```


## Replay
**Replaying was not run**.

