package spinner

import (
	"fmt"
	"time"
)

type Spinner struct {
    stopChan chan struct{}
}

func NewSpinner() *Spinner {
    return &Spinner{
        stopChan: make(chan struct{}),
    }
}

func (s *Spinner) Start(Message string) {
    go func() {
        for {
            select {
            case <-s.stopChan:
                return
            default:
                fmt.Printf("\r🔆 %s.   ", Message)
                time.Sleep(100 * time.Millisecond)
                fmt.Printf("\r🔅 %s..  ", Message)
                time.Sleep(100 * time.Millisecond)
                fmt.Printf("\r🔆 %s... ", Message)
                time.Sleep(100 * time.Millisecond)
                fmt.Printf("\r🔅 %s....", Message)
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
}

func (s *Spinner) Stop(result bool) {
    if result {
        fmt.Print("\033[2K\r✔️ Success\n")
    } else {
        fmt.Print("\033[2K\r❌ Failure\n")
    }
    close(s.stopChan)
}
