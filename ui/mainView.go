package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MainView struct {
	App      *App
	name     string
	root     *tview.Flex
	grid     *tview.Grid
	urlsList *tview.List
	active   bool
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

func NewMainView(app *App) *MainView {
	mainView := &MainView{
		App:      app,
		name:     "MainView",
		root:     tview.NewFlex(),
		grid:     tview.NewGrid(),
		urlsList: tview.NewList(),
		active:   true,
		stopCh:   make(chan struct{}),
	}

	mainView.urlsList.ShowSecondaryText(false)
	mainView.urlsList.SetHighlightFullLine(true)
	mainView.urlsList.SetMainTextColor(tcell.ColorBlue)
	mainView.grid.SetBorder(true)
	mainView.grid.AddItem(mainView.urlsList, 0, 0, 1, 1, 0, 0, true)
	mainView.root.SetDirection(tview.FlexRow).AddItem(mainView.grid, 0, 1, true)

	return mainView
}

func (m *MainView) IsActive() bool {
	return m.active
}

func (m *MainView) SetActive(status bool) {
	m.active = status
}

func (m *MainView) Name() string {
	return m.name
}

func (m *MainView) Root() tview.Primitive {
	return m.root
}

func (m *MainView) SetupEvents() {
	m.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			m.App.SwitchToPage("ConfirmQuitView")
		} else if event.Rune() == 'a' {
			m.App.AddItem()
		} else if event.Rune() == 'd' {
			m.App.RemoveItem()
		} else if event.Rune() == 'f' {
			m.App.SortByComplete()
		} else if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		} else if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		} else if event.Rune() == '/' {
			searchView := m.App.views["SearchView"].(*SearchView)
			searchView.input.SetText("")
			searchView.results.Clear()

			for _, item := range m.App.urls {
				searchView.results.AddItem(item.Url, "", 0, nil)
			}

			m.App.DisplayPage("SearchView")
		} else if event.Key() == tcell.KeyEnter {
			if m.urlsList.GetItemCount() == 0 {
				return event
			}
			index := m.urlsList.GetCurrentItem()

			logsView := m.App.views["LogsView"].(*LogsView)
			logsView.setLogText(m.App.urls[index])
			m.App.SwitchToPage("LogsView")
		}
		return event
	})

}
