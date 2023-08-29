package ui

import (
	"fmt"
	"time"

	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ViewController interface {
	IsActive() bool
	SetActive(bool)
}

type App struct {
	*tview.Application
	pages       *tview.Pages
	views       map[string]ViewController
	currentView string
	urls        []*url.UrlItem
}

func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
		views:       make(map[string]ViewController),
		currentView: "MainView",
		urls:        []*url.UrlItem{},
	}

	mainView := newMainView()
	logsView := newLogsView()
	urlFormView := newUrlFormView()

	app.pages.AddPage("MainView", mainView.root, true, true)
	app.pages.AddPage("LogsView", logsView.root, true, false)
	app.pages.AddPage("UrlFormView", urlFormView.root, true, false)

	app.views["MainView"] = mainView
	app.views["LogsView"] = logsView
	app.views["UrlFormView"] = urlFormView

	logsView.log.SetChangedFunc(func() {
		app.Draw()
	})
	logsView.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.SwitchToPage("MainView")
		}
		return event
	})

	mainView.urlsList.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		app.SwitchToPage("LogsView")
		logsView.setLogText(app.urls[index])
	})
	mainView.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'a' {
			app.AddItem()
		} else if event.Rune() == 'd' {
			app.RemoveItem()
		}
		return event
	})

	go func() {
		for {
			app.Draw()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	app.SetRoot(app.pages, true)
	app.EnableMouse(true)

	return app
}

func (a *App) SwitchToPage(page string) {
	a.views[a.currentView].SetActive(false)
	a.currentView = page
	a.views[a.currentView].SetActive(true)
	a.pages.SwitchToPage(page)
}

func (a *App) AddItem() {
	item := &url.UrlItem{}
	urlFormView := a.views["UrlFormView"].(*UrlFormView)

	okAction := func() {
		item.Start()
		a.urls = append(a.urls, item)
		a.RedrawList()
		a.SwitchToPage("MainView")
	}
	cancelAction := func() {
		a.SwitchToPage("MainView")
	}
	urlFormView.contructForm(item, okAction, cancelAction)

	a.SwitchToPage("UrlFormView")
}

func (a *App) ItemStatusUpdater(item *url.UrlItem, itemIdx int) {
	mainView := a.views["MainView"].(*MainView)

	go func() {
		defer mainView.wg.Done()

		for {
			select {
			case <-mainView.stopCh:
				return
			default:
				if item.Recording {
					mainView.urlsList.SetItemText(itemIdx, fmt.Sprintf("%s ([green]recording[blue])", item.Url), "")
				} else {
					mainView.urlsList.SetItemText(itemIdx, fmt.Sprintf("%s ([red]done[blue])", item.Url), "")
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()
}

func (a *App) RedrawList() {
	mainView := a.views["MainView"].(*MainView)

	close(mainView.stopCh)
	mainView.wg.Wait()

	mainView.stopCh = make(chan struct{})
	mainView.wg.Add(len(a.urls))

	mainView.urlsList.Clear()
	for idx, item := range a.urls {
		mainView.urlsList.AddItem(item.Url, "", 0, nil)
		a.ItemStatusUpdater(item, idx)
	}
}

func (a *App) RemoveItem() {
	if len(a.urls) == 0 {
		return
	}

	mainView := a.views["MainView"].(*MainView)

	curr := mainView.urlsList.GetCurrentItem()
	go a.urls[curr].Stop()
	a.urls = append(a.urls[:curr], a.urls[curr+1:]...)

	a.RedrawList()
}

func (a *App) CleanUp() {
	for _, item := range a.urls {
		item.Stop()
	}
}
