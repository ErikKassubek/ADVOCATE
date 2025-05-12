package gopie

import (
	"sync"
	"time"
)

// This program is an example, of how the fully partial replay in GoPie
// may miss the execution of the mutated section, while the partially full replay
// in GoPie+ and GoPieHB can avoid this.
// Lets assume in the recorded run, the TryLock was executed before to Lock,
// therefore being able to acquire the mutex. Let's also assume, that the
// mutated code position is fully in the mutatedCode function.
// If we let the code run freely and only enforce the order in the mutated
// section, it is possible that in execution of this run, the Lock happens
// before the TryLock and the TryLock is not able to acquire the lock.
// In this case, the program would never get to the relevant, mutated section.
// By enforcing the exact order of the code before the relevant section,
// like it is done in GoPie+ and GoPieHB, we can avoid this, since the replay
// mechanism makes sure, that the Lock and TryLock are executed in the
// correct order.

func main() {
	m := sync.Mutex{}

	go func() {
		m.Lock()
		// some code
		time.Sleep(time.Second)
		m.Unlock()
	}()

	res := m.TryLock()
	if res {
		mutatedCode()
		m.Unlock()
	}

}
