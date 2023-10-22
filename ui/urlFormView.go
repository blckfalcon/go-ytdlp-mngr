package ui

import (
	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/rivo/tview"
)

type UrlFormView struct {
	root   *tview.Form
	active bool
}

func NewUrlFormView() *UrlFormView {
	urlFormView := &UrlFormView{
		root:   tview.NewForm(),
		active: false,
	}

	urlFormView.root.SetBorder(true)

	return urlFormView
}

func (u *UrlFormView) contructForm(item *url.UrlItem, okAction func(), cancelAction func()) {
	u.root.Clear(true)

	u.root.AddInputField("Url", "", 256, nil, func(url string) {
		item.Url = url
	})

	u.root.AddButton("Save", func() { okAction() })
	u.root.AddButton("Cancel", func() { cancelAction() })
}

func (u *UrlFormView) IsActive() bool {
	return u.active
}

func (u *UrlFormView) SetActive(status bool) {
	u.active = status
}
