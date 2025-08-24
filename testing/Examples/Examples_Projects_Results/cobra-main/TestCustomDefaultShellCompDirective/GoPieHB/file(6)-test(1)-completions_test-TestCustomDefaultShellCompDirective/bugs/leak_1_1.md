# Leak: Leak on routine

The analyzer detected a leak on a routine without a blocking operations.
This means that the routine was terminated because of a panic in another routine or because the main routine terminated while this routine was still running.
This can be a desired behavior, but it can also be a signal for a not otherwise detected block.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestCustomDefaultShellCompDirective
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/cobra-main/completions_test.go

## Bug Elements
The elements involved in the found leak are located at the following positions:

###  Routine: End
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/cobra-main/completions.go:569
```go
558 ...
559 
560 			// This is for commands that have subcommands but also specify a ValidArgsFunction.
561 		}
562 	}
563 
564 	// Find the completion function for the flag or command
565 	var completionFn CompletionFunc
566 	if flag != nil && flagCompletion {
567 		flagCompletionMutex.RLock()
568 		completionFn = flagCompletionFunctions[flag]
569 		flagCompletionMutex.RUnlock()           // <-------
570 	} else {
571 		completionFn = finalCmd.ValidArgsFunction
572 	}
573 	if completionFn != nil {
574 		// Go custom completion defined for this flag or command.
575 		// Call the registered completion function to get the completions.
576 		var comps []Completion
577 		comps, directive = completionFn(finalCmd, finalArgs, toComplete)
578 		completions = append(completions, comps...)
579 	}
580 
581 ...
```


## Replay
**Replaying was not run**.

