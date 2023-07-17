package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UrlItem struct {
	url       string
	cmd       *exec.Cmd
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	doneCh    chan error
	recording bool
}

func (u *UrlItem) Stop() {
	var err error
	if u.cmd.ProcessState != nil && u.cmd.ProcessState.Exited() {
		u.recording = false
		return
	}

	err = u.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		log.Println(err)
	}

	startTime := time.Now()
	for u.recording {
		elapsed := time.Since(startTime)
		if elapsed > 30*time.Second {
			fmt.Println("Timeout occurred")
			break
		}
	}
	u.recording = false
}

var currentView string
var urls []*UrlItem

func addUrlItemForm(page *tview.Pages, form *tview.Form, list *tview.List) {
	item := UrlItem{}

	form.AddInputField("Url", "", 256, nil, func(url string) {
		item.url = url
	})

	form.AddButton("Save", func() {
		item.cmd = exec.Command("yt-dlp", item.url)
		item.stdout, _ = item.cmd.StdoutPipe()
		item.stderr, _ = item.cmd.StderrPipe()

		err := item.cmd.Start()
		if err != nil {
			return
		}

		item.recording = true
		item.doneCh = make(chan error, 1)

		go func(i *UrlItem) {
			i.doneCh <- i.cmd.Wait()
		}(&item)

		go func(i *UrlItem) {
			<-i.doneCh
			i.recording = false
			i.stdout.Close()
			i.stderr.Close()
		}(&item)

		urls = append(urls, &item)
		addUrlToList(list)

		currentView = "MainView"
		page.SwitchToPage(currentView)
	})
}

func addUrlToList(list *tview.List) {
	list.Clear()
	for _, item := range urls {
		list.AddItem(item.url, "", 0, nil)
	}
}

func setLogText(log *tview.TextView, item *UrlItem) {
	log.Clear()
	if !item.recording {
		log.SetText("yt-dlp is done")
		return
	}

	builder1 := strings.Builder{}
	builder2 := strings.Builder{}
	builder3 := strings.Builder{}
	donePipeInLog := make(chan bool, 1)
	donePipeErrLog := make(chan bool, 1)
	go readFromPipe(donePipeInLog, item.stdout, &builder1)
	go readFromPipe(donePipeErrLog, item.stderr, &builder2)

	go func() {
		for {
			if !item.recording || currentView != "LogsView" {
				donePipeInLog <- true
				donePipeErrLog <- true
				return
			}
			builder3.Reset()
			builder3.WriteString(builder1.String())
			builder3.WriteString(builder2.String())
			l := builder3.String()
			log.SetText(l)
			time.Sleep(50 * time.Millisecond)
		}
	}()
}

func readFromPipe(done chan bool, pipe io.Reader, writer io.Writer) {
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

func main() {
	var app = tview.NewApplication()
	var pages = tview.NewPages()
	var flex = tview.NewFlex()
	var urlsList = tview.NewList().ShowSecondaryText(false)
	var urlform = tview.NewForm()
	var logs = tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})
	var listView = tview.NewGrid()
	var logsView = tview.NewGrid()

	urlform.SetBorder(true)
	listView.SetBorder(true)
	logsView.SetBorder(true)
	logsView.SetBorders(true).SetRows(1, 5)
	logsView.SetBorderPadding(-1, -1, -1, -1)

	pages.AddPage("MainView", flex, true, true)
	pages.AddPage("LogsView", logsView, true, false)
	pages.AddPage("UrlForm", urlform, true, false)

	listView.AddItem(urlsList, 0, 0, 1, 1, 0, 0, true)
	urlsList.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		currentView = "LogsView"
		pages.SwitchToPage(currentView)
		setLogText(logs, urls[index])
	})

	logsView.AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Stdout"), 0, 0, 1, 1, 0, 0, false)
	logsView.AddItem(logs, 1, 0, 2, 1, 0, 0, true)
	logsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			currentView = "MainView"
			pages.SwitchToPage(currentView)
		}
		return event
	})

	flex.SetDirection(tview.FlexRow).AddItem(listView, 0, 1, true)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'a' {
			urlform.Clear(true)
			addUrlItemForm(pages, urlform, urlsList)
			urlform.AddButton("Cancel", func() {
				currentView = "MainView"
				pages.SwitchToPage(currentView)
			})
			currentView = "UrlForm"
			pages.SwitchToPage(currentView)
		} else if event.Rune() == 'd' {
			curr := urlsList.GetCurrentItem()
			go urls[curr].Stop()
			urlsList.RemoveItem(curr)
			urls = append(urls[:curr], urls[curr+1:]...)
		}
		return event
	})

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

	for _, item := range urls {
		item.Stop()
	}
}
