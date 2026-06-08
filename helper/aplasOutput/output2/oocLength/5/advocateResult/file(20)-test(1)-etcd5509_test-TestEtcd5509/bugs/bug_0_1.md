# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestEtcd5509
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/5509/etcd5509_test.go
- Trace: advocateTrace_21

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/5509/etcd5509_test.go#27
```go
16 ...
17 
18 func (c *Client) Close() {
19 	c.mu.Lock()
20 	defer c.mu.Unlock()
21 	if c.cancel == nil {
22 		return
23 	}
24 	c.cancel()
25 	c.cancel = nil
26 	c.mu.Unlock()
27 	c.mu.Lock() // block here           // <-------
28 }
29 
30 type remoteClient struct {
31 	client *Client
32 	mu     sync.Mutex
33 }
34 
35 func (r *remoteClient) acquire(ctx context.Context) error {
36 	for {
37 		r.client.mu.RLock()
38 
39 ...
```


###  Channel: Close
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/5509/etcd5509_test.go#99
```go
88 ...
89 
90 	}
91 	kv := NewKV(cli)
92 	donec := make(chan struct{})
93 	go func() {
94 		defer close(donec)
95 		err := kv.Get(context.TODO())
96 		if err != nil && err != ErrConnClosed {
97 			fmt.Println("Expect ErrConnClosed")
98 		}
99 	}()           // <-------
100 
101 	cli.Close()
102 
103 	<-donec
104 }
105 
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/5509/etcd5509_test.go#37
```go
26 ...
27 
28 }
29 
30 type remoteClient struct {
31 	client *Client
32 	mu     sync.Mutex
33 }
34 
35 func (r *remoteClient) acquire(ctx context.Context) error {
36 	for {
37 		r.client.mu.RLock()           // <-------
38 		closed := r.client.cancel == nil
39 		r.mu.Lock()
40 		r.mu.Unlock()
41 		if closed {
42 			return ErrConnClosed // Missing RUnlock before return
43 		}
44 		r.client.mu.RUnlock()
45 	}
46 }
47 
48 
49 ...
```


-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/5/blocking/etcd/5509/etcd5509_test.go#103
```go
92 ...
93 
94 		defer close(donec)
95 		err := kv.Get(context.TODO())
96 		if err != nil && err != ErrConnClosed {
97 			fmt.Println("Expect ErrConnClosed")
98 		}
99 	}()
100 
101 	cli.Close()
102 
103 	<-donec           // <-------
104 }
105 
```


## Replay
**Replaying was not run**.

