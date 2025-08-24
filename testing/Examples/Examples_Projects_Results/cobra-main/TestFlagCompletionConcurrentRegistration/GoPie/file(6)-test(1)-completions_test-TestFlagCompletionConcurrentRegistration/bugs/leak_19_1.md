# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestFlagCompletionConcurrentRegistration
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/cobra-main/completions_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/cobra-main/completions_test.go:2145
```go
2134 ...
2135 
2136 	}
2137 
2138 	rootCmd.AddCommand(childCmd)
2139 
2140 	// Register completion in different threads to test concurrency.
2141 	var wg sync.WaitGroup
2142 	for i := 1; i <= maxFlags; i++ {
2143 		index := i
2144 		flagName := fmt.Sprintf("flag%d", i)
2145 		wg.Add(1)           // <-------
2146 		go func() {
2147 			defer wg.Done()
2148 			cmd := rootCmd
2149 			if index%2 == 0 {
2150 				cmd = childCmd
2151 			}
2152 			_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
2153 				return []string{fmt.Sprintf("flag%d", index)}, ShellCompDirectiveDefault
2154 			})
2155 		}()
2156 
2157 ...
```


## Replay
**Replaying was not run**.

