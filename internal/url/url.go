package url

import (
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type DownloadStage int

const (
	StageNotStarted DownloadStage = iota
	StageDownloading
	StageProcessing
	StageCompleted
	StageError
)

func (s DownloadStage) String() string {
	stages := [...]string{
		"Not Started",
		"Downloading",
		"Processing",
		"Completed",
		"Error",
	}

	if s < StageNotStarted || s > StageError {
		return "Unknown"
	}

	return stages[s]
}

type UrlItem struct {
	Url       string
	cmd       *exec.Cmd
	Stdout    io.ReadCloser
	StdoutBuf chan []byte
	Stderr    io.ReadCloser
	StderrBuf chan []byte
	Recording DownloadStage
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

	u.Recording = StageDownloading
	u.Logging = false
	var wg sync.WaitGroup
	wg.Add(2)

	go func(i *UrlItem) {
		_ = i.cmd.Wait()

		i.Recording = StageProcessing
		wg.Wait()
		close(i.StdoutBuf)
		close(i.StderrBuf)

		if i.cmd.ProcessState != nil && i.cmd.ProcessState.Exited() {
			i.Recording = StageCompleted
		}
	}(u)

	sendReadToBuffer := func(bufCh chan<- []byte, reader io.Reader) {
		buffer := make([]byte, 1000)
		defer wg.Done()
		for u.Recording == StageDownloading {
			_, _ = reader.Read(buffer)
			if u.Logging {
				bufCh <- buffer
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	go sendReadToBuffer(u.StdoutBuf, u.Stdout)
	go sendReadToBuffer(u.StderrBuf, u.Stderr)
}

func (u *UrlItem) Stop() {
	var err error
	if u.cmd.ProcessState != nil && u.cmd.ProcessState.Exited() {
		u.Recording = StageCompleted
		return
	}

	err = u.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		log.Println(err)
	}

	backoff := 1 * time.Second
	maxBackoff := 30 * time.Minute
	for u.Recording == StageProcessing {
		if u.cmd.ProcessState != nil && u.cmd.ProcessState.Exited() {
			u.Recording = StageCompleted
			return
		}

		if backoff < maxBackoff {
			backoff *= 2
		} else {
			u.Recording = StageCompleted
			return
		}
		time.Sleep(backoff)
	}
}
