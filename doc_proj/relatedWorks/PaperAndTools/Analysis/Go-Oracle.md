# Go-Oracle: Automated Test Oracle for Go Concurrency Bugs

[F. Tsimpourlas, C. Peng, C. Rosuero, P. Yang, and A. Rajan, Go-oracle: Automated test oracle for go concurrency bugs, 2024. arXiv: 2412 . 08061 [cs.SE]. [Online]. Available:https://arxiv.org/abs/2412.08061.](https://arxiv.org/abs/2412.08061)

## Summary

The paper addresses the challenge of detecting and classifying concurrency bugs in Go programs. Go's concurrency model, which combines communicating sequential processes with shared memory, presents unique difficulties in identifying such bugs. To tackle this, the authors developed Go-Oracle, an automated test oracle designed to classify Go program executions as passing or failing, specifically focusing on concurrency issues. ​

The approach involves collecting execution traces from Go programs using Go's native execution tracer, capturing events like goroutine creation, synchronization operations, and system calls. These traces are then preprocessed and encoded to train a transformer-based neural network capable of distinguishing between passing and failing executions.