package ui

import (
	"fmt"
	"sort"
	"time"

	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ViewController interface {
	IsActive() bool
	SetActive(bool)
	Name() string
	Root() tview.Primitive
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

	mainView := NewMainView()
	logsView := NewLogsView()
	urlFormView := NewUrlFormView()
	confirmQuitView := NewConfirmQuitView()

	app.AddView(mainView, true, true)
	app.AddView(logsView, true, false)
	app.AddView(urlFormView, true, false)
	app.AddView(confirmQuitView, false, false)

	logsView.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.SwitchToPage("MainView")
		}
		return event
	})

	confirmQuitView.root.SetDoneFunc(func(_ int, buttonLabel string) {
		if buttonLabel == "Quit" {
			app.Stop()
		}
		app.SwitchToPage("MainView")
	})

	mainView.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.SwitchToPage("ConfirmQuitView")
		} else if event.Rune() == 'a' {
			app.AddItem()
		} else if event.Rune() == 'd' {
			app.RemoveItem()
		} else if event.Rune() == 'f' {
			app.SortByComplete()
		} else if event.Key() == tcell.KeyEnter {
			if mainView.urlsList.GetItemCount() == 0 {
				return event
			}
			index := mainView.urlsList.GetCurrentItem()
			app.SwitchToPage("LogsView")
			logsView.setLogText(app.urls[index])
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

func (a *App) AddView(view ViewController, resize bool, visible bool) {
	a.pages.AddPage(view.Name(), view.Root(), resize, visible)
	a.views[view.Name()] = view
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
				var recordStatus string
				switch item.Recording {
				case url.StageNotStarted:
					recordStatus = "blue"
				case url.StageDownloading:
					recordStatus = "green"
				case url.StageCompleted:
					recordStatus = "darkcyan"
				case url.StageProcessing:
					recordStatus = "magenta"
				case url.StageError:
					recordStatus = "red"
				}

				mainView.urlsList.SetItemText(
					itemIdx,
					fmt.Sprintf("%-50s [blue]([%s]%s[blue])", item.Url, recordStatus, item.Recording),
					"",
				)
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

	curr := mainView.urlsList.GetCurrentItem()

	mainView.urlsList.Clear()
	for idx, item := range a.urls {
		mainView.urlsList.AddItem(item.Url, "", 0, nil)
		a.ItemStatusUpdater(item, idx)
	}

	mainView.urlsList.SetCurrentItem(curr)
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

func (a *App) SortByComplete() {
	sort.Sort(url.ByComplete(a.urls))
	a.RedrawList()
}

func (a *App) CleanUp() {
	for _, item := range a.urls {
		item.Stop()
	}
}
