package ui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MainView struct {
	name     string
	root     *tview.Flex
	grid     *tview.Grid
	urlsList *tview.List
	active   bool
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

func NewMainView() *MainView {
	mainView := &MainView{
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
