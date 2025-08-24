# Bug: Unknown Panic

The execution of the program timed out

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestTreeInvalidNodeType
- File: /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree_test.go

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Unknown element type
-> /home/ilian/Projects/Go_Fuzzing_Concurrency/Examples/Examples_Projects/gin-master/tree.go:592
```go
581 ...
582 
583 							Value: val,
584 						}
585 					}
586 
587 					value.handlers = n.handlers
588 					value.fullPath = n.fullPath
589 					return value
590 
591 				default:
592 					panic("invalid node type")           // <-------
593 				}
594 			}
595 		}
596 
597 		if path == prefix {
598 			// If the current path does not equal '/' and the node does not have a registered handle and the most recently matched node has a child node
599 			// the current node needs to roll back to last valid skippedNode
600 			if n.handlers == nil && path != "/" {
601 				for length := len(*skippedNodes); length > 0; length-- {
602 					skippedNode := (*skippedNodes)[length-1]
603 
604 ...
```


## Replay
**Replaying was not run**.

