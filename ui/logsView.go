package ui

import (
	"bufio"
	"bytes"
	"io"
	"time"

	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/rivo/tview"
)

type LogsView struct {
	root   *tview.Grid
	title  *tview.TextView
	log    *tview.TextView
	active bool
}

func newLogsView() *LogsView {
	logsView := &LogsView{
		root:   tview.NewGrid(),
		title:  tview.NewTextView(),
		log:    tview.NewTextView(),
		active: false,
	}

	logsView.title.SetTextAlign(tview.AlignCenter).SetText("Stdout")
	logsView.log.SetMaxLines(1000)

	logsView.root.SetBorder(true)
	logsView.root.SetBorders(true).SetRows(1, 5)
	logsView.root.SetBorderPadding(-1, -1, -1, -1)

	logsView.root.AddItem(logsView.title, 0, 0, 1, 1, 0, 0, false)
	logsView.root.AddItem(logsView.log, 1, 0, 2, 1, 0, 0, true)

	return logsView
}

func (l *LogsView) setLogText(item *url.UrlItem) {
	if !item.Recording {
		l.SetLogMessage("yt-dlp is done")
		return
	}

	buffer1 := bytes.Buffer{}
	buffer2 := bytes.Buffer{}
	donePipeInLog := make(chan bool, 1)
	donePipeErrLog := make(chan bool, 1)

	readFromPipe := func(done <-chan bool, pipe io.Reader, writer io.Writer) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			select {
			case <-done:
				return
			default:
				line := scanner.Text() + "\n"
				_, err := writer.Write([]byte(line))
				if err != nil {
					return
				}
			}
		}
	}
	go readFromPipe(donePipeInLog, item.Stdout, &buffer1)
	go readFromPipe(donePipeErrLog, item.Stderr, &buffer2)

	go func() {
		for {
			if !item.Recording || !l.active {
				donePipeInLog <- true
				donePipeErrLog <- true
				l.SetLogMessage("yt-dlp is done")
				return
			}
			log := buffer1.String() + buffer2.String()
			l.log.SetText(log)
			time.Sleep(200 * time.Millisecond)
		}
	}()
}

func (l *LogsView) SetLogMessage(msg string) {
	l.log.Clear()
	l.log.SetText(msg)
}

func (l *LogsView) IsActive() bool {
	return l.active
}

func (l *LogsView) SetActive(status bool) {
	l.active = status
}
