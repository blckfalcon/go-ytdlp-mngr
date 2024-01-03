package ui

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"time"

	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/rivo/tview"
)

type LogsView struct {
	name   string
	root   *tview.Grid
	title  *tview.TextView
	log    *tview.TextView
	active bool
}

func NewLogsView() *LogsView {
	logsView := &LogsView{
		name:   "LogsView",
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
	if item.Recording != url.StageDownloading {
		l.SetLogMessage("yt-dlp is done")
		return
	}

	item.Logging = true
	buffer1 := bytes.Buffer{}
	buffer2 := bytes.Buffer{}
	donePipeInLog := make(chan bool, 1)
	donePipeErrLog := make(chan bool, 1)

	readFromPipe := func(done <-chan bool, data chan []byte, writer io.Writer) {
		for {
			select {
			case <-done:
				return
			case d := <-data:
				r := bytes.NewReader(d)
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					ansiRe := regexp.MustCompile(`\x1b[^m]*m`)
					controlRe := regexp.MustCompile(`[\x00-\x1f\x7f]`)
					line := scanner.Text()
					line = ansiRe.ReplaceAllString(line, "")
					line = controlRe.ReplaceAllString(line, "")
					line = line + "\n"
					_, err := writer.Write([]byte(line))
					if err != nil {
						return
					}
				}
			default:
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
	go readFromPipe(donePipeInLog, item.StdoutBuf, &buffer1)
	go readFromPipe(donePipeErrLog, item.StderrBuf, &buffer2)

	go func() {
		for {
			if item.Recording != url.StageDownloading || !l.active {
				donePipeInLog <- true
				donePipeErrLog <- true
				close(donePipeInLog)
				close(donePipeErrLog)
				item.Logging = false
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

func (l *LogsView) Name() string {
	return l.name
}

func (l *LogsView) Root() tview.Primitive {
	return l.root
}
