package aof

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/Fashad-Ahmed/fluxkv/internal/resp"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}


func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// Optional but highly recommended: Start a background goroutine to sync 
	// the file to disk every second (just like Redis does!)
	go func() {
		for {
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (a *Aof) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

// Write safely appends a RESP command to the file
func (a *Aof) Write(value resp.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.file.Write(value.Marshal())
	if err != nil {
		return err
	}
	return nil
}

// Read loops through the file and passes each parsed command to a callback function
func (a *Aof) Read(callback func(value resp.Value)) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Move the file cursor back to the beginning
	a.file.Seek(0, 0)
	
	// Re-use our exact same TCP parser, but point it at the file instead!
	reader := resp.NewResp(a.file)

	for {
		value, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		callback(value)
	}
	return nil
}