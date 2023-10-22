package ui

import (
	"github.com/rivo/tview"
)

type ConfirmQuitView struct {
	root   *tview.Modal
	active bool
}

func NewConfirmQuitView() *ConfirmQuitView {
	confirmQuitView := &ConfirmQuitView{
		root:   tview.NewModal(),
		active: false,
	}

	confirmQuitView.root.SetText("Do you want to quit the application?")
	confirmQuitView.root.AddButtons([]string{"Quit", "Cancel"})

	return confirmQuitView
}

func (c *ConfirmQuitView) IsActive() bool {
	return c.active
}

func (c *ConfirmQuitView) SetActive(status bool) {
	c.active = status
}
