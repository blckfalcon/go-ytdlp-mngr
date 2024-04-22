package ui

import (
	"github.com/rivo/tview"
)

type ConfirmQuitView struct {
	App    *App
	name   string
	root   *tview.Modal
	active bool
}

func NewConfirmQuitView(app *App) *ConfirmQuitView {
	confirmQuitView := &ConfirmQuitView{
		App:    app,
		name:   "ConfirmQuitView",
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

func (c *ConfirmQuitView) Name() string {
	return c.name
}

func (c *ConfirmQuitView) Root() tview.Primitive {
	return c.root
}

func (c *ConfirmQuitView) SetupEvents() {
	c.root.SetDoneFunc(func(_ int, buttonLabel string) {
		if buttonLabel == "Quit" {
			c.App.Stop()
		}
		c.App.SwitchToPage("MainView")
	})
}
