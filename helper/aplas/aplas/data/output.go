package data

import (
	"os"
	"path/filepath"
	"sync"
)

var OutputLock = sync.Mutex{}

func CreateOutput(name ...string) (string, error) {
	OutputLock.Lock()
	defer OutputLock.Unlock()

	path, err := filepath.Abs(filepath.Join(PathOutput, filepath.Join(name...)))
	if err != nil {
		return path, err
	}

	err = os.Mkdir(path, 0755)

	if err != nil {
		return path, err
	}

	return path, err
}
