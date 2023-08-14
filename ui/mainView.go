package ui

import "github.com/rivo/tview"

type MainView struct {
	root     *tview.Flex
	grid     *tview.Grid
	urlsList *tview.List
	active   bool
}

func newMainView() *MainView {
	mainView := &MainView{
		root:     tview.NewFlex(),
		grid:     tview.NewGrid(),
		urlsList: tview.NewList(),
		active:   true,
	}

	mainView.urlsList.ShowSecondaryText(false)
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
