package url

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"syscall"
	"time"
)

type UrlItem struct {
	Url       string
	cmd       *exec.Cmd
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser
	doneCh    chan error
	Recording bool
}

func (u *UrlItem) Start() {
	u.cmd = exec.Command("yt-dlp", u.Url)
	u.Stdout, _ = u.cmd.StdoutPipe()
	u.Stderr, _ = u.cmd.StderrPipe()

	err := u.cmd.Start()
	if err != nil {
		return
	}

	u.Recording = true
	u.doneCh = make(chan error, 1)

	go func(i *UrlItem) {
		i.doneCh <- i.cmd.Wait()
	}(u)

	go func(i *UrlItem) {
		<-i.doneCh
		i.Recording = false
		i.Stdout.Close()
		i.Stderr.Close()
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
