package url

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type UrlItem struct {
	Url       string
	cmd       *exec.Cmd
	Stdout    io.ReadCloser
	StdoutBuf chan []byte
	Stderr    io.ReadCloser
	StderrBuf chan []byte
	Recording bool
	Logging   bool
}

func (u *UrlItem) Start() {
	u.cmd = exec.Command("yt-dlp", u.Url)
	u.Stdout, _ = u.cmd.StdoutPipe()
	u.Stderr, _ = u.cmd.StderrPipe()

	u.StdoutBuf = make(chan []byte, 1)
	u.StderrBuf = make(chan []byte, 1)

	err := u.cmd.Start()
	if err != nil {
		return
	}

	u.Recording = true
	u.Logging = false
	var wg sync.WaitGroup
	wg.Add(2)

	go func(i *UrlItem) {
		_ = i.cmd.Wait()

		i.Recording = false
		wg.Wait()
		close(i.StdoutBuf)
		close(i.StderrBuf)
	}(u)

	go func(i *UrlItem) {
		buffer := make([]byte, 1000)
		defer wg.Done()
		for i.Recording {
			_, _ = i.Stdout.Read(buffer)
			if i.Logging {
				i.StdoutBuf <- buffer
			}
			time.Sleep(200 * time.Millisecond)
		}
	}(u)

	go func(i *UrlItem) {
		buffer := make([]byte, 1000)
		defer wg.Done()
		for i.Recording {
			_, _ = i.Stderr.Read(buffer)
			if i.Logging {
				i.StderrBuf <- buffer
			}
			time.Sleep(200 * time.Millisecond)
		}
	}(u)
}

func (u *UrlItem) Stop() {
	var err error
	if u.cmd.ProcessState != nil && u.cmd.ProcessState.Exited() {
		u.Recording = false
		return
	}

	err = u.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		log.Println(err)
	}

	startTime := time.Now()
	for u.Recording {
		elapsed := time.Since(startTime)
		if elapsed > 1*time.Minute {
			fmt.Println("Timeout occurred")
			break
		}
	}
	u.Recording = false
}
