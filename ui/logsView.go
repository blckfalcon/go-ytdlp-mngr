package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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

	logsView.root.SetBorder(true)
	logsView.root.SetBorders(true).SetRows(1, 5)
	logsView.root.SetBorderPadding(-1, -1, -1, -1)

	logsView.root.AddItem(logsView.title, 0, 0, 1, 1, 0, 0, false)
	logsView.root.AddItem(logsView.log, 1, 0, 2, 1, 0, 0, true)

	return logsView
}

func (l *LogsView) setLogText(item *url.UrlItem) {
	l.log.Clear()
	if !item.Recording {
		l.log.SetText("yt-dlp is done")
		return
	}

	builder1 := strings.Builder{}
	builder2 := strings.Builder{}
	builder3 := strings.Builder{}
	donePipeInLog := make(chan bool, 1)
	donePipeErrLog := make(chan bool, 1)

	readFromPipe := func(done chan bool, pipe io.Reader, writer io.Writer) {
		scanner := bufio.NewScanner(pipe)
		for {
			select {
			case <-done:
				return
			default:
				for scanner.Scan() {
					fmt.Fprintln(writer, scanner.Text())
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
	}
	go readFromPipe(donePipeInLog, item.Stdout, &builder1)
	go readFromPipe(donePipeErrLog, item.Stderr, &builder2)

	go func() {
		for {
			if !item.Recording || !l.active {
				donePipeInLog <- true
				donePipeErrLog <- true
				return
			}
			builder3.Reset()
			builder3.WriteString(builder1.String())
			builder3.WriteString(builder2.String())
			log := builder3.String()
			l.log.SetText(log)
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func (l *LogsView) IsActive() bool {
	return l.active
}

func (l *LogsView) SetActive(status bool) {
	l.active = status
}
