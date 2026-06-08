# Bug: P05 - Possible Cyclic Deadlock

The analysis detected a possible cyclic deadlock.
If this deadlock contains or influences the run of the main routine, this can result in the program getting stuck. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMoby4951
- File: /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/4951/moby4951_test.go
- Trace: advocateTrace_65

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/4951/moby4951_test.go#33
```go
22 ...
23 
24 	nrDeletedDevices int
25 }
26 
27 func (devices *DeviceSet) DeleteDevice(hash string) {
28 	devices.Lock()
29 	defer devices.Unlock()
30 
31 	info := devices.lookupDevice(hash)
32 
33 	info.lock.Lock()           // <-------
34 	defer info.lock.Unlock()
35 
36 	devices.deleteDevice(info)
37 }
38 
39 func (devices *DeviceSet) lookupDevice(hash string) *DevInfo {
40 	existing, ok := devices.infos[hash]
41 	if !ok {
42 		return nil
43 	}
44 
45 ...
```


###  Mutex: Part of deadlock
-> /home/advocate/Advocate/Experiments/Advocate/helper/aplas/output/oocLength/10/blocking/moby/4951/moby4951_test.go#55
```go
44 ...
45 
46 
47 func (devices *DeviceSet) deleteDevice(info *DevInfo) {
48 	devices.removeDeviceAndWait(info.Name())
49 }
50 
51 func (devices *DeviceSet) removeDeviceAndWait(devname string) {
52 	/// remove devices by devname
53 	devices.Unlock()
54 	time.Sleep(300 * time.Nanosecond)
55 	devices.Lock()           // <-------
56 }
57 
58 type DevInfo struct {
59 	lock sync.Mutex
60 	name string
61 }
62 
63 func (info *DevInfo) Name() string {
64 	return info.name
65 }
66 
67 ...
```


## Replay
**Replaying was not run**.

