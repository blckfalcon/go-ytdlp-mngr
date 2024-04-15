package ui

import (
	"github.com/blckfalcon/go-ytdlp-mngr/internal/url"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SearchView struct {
	name    string
	root    *tview.Grid
	title   *tview.TextView
	input   *tview.InputField
	results *tview.List
	urls    []*url.UrlItem
	active  bool
}

func NewSearchView() *SearchView {
	searchView := &SearchView{
		name:    "SearchView",
		root:    tview.NewGrid(),
		title:   tview.NewTextView(),
		input:   tview.NewInputField(),
		results: tview.NewList(),
		urls:    []*url.UrlItem{},
		active:  false,
	}

	searchView.title.SetTextAlign(tview.AlignCenter).SetText("Fuzzy search items")

	searchView.results.ShowSecondaryText(false)
	searchView.results.SetHighlightFullLine(true)
	searchView.results.SetMainTextColor(tcell.ColorBlue)

	searchView.root.SetBorder(true)
	searchView.root.SetBorders(true)
	searchView.root.SetColumns(-1, 100, -1)
	searchView.root.SetRows(-1, 1, 1, 15, -1)

	searchView.root.AddItem(searchView.title, 1, 1, 1, 1, 0, 0, false)
	searchView.root.AddItem(searchView.input, 2, 1, 1, 1, 0, 0, true)
	searchView.root.AddItem(searchView.results, 3, 1, 1, 1, 0, 0, false)

	return searchView
}

func (s *SearchView) IsActive() bool {
	return s.active
}

func (s *SearchView) SetActive(status bool) {
	s.active = status
}

func (s *SearchView) Name() string {
	return s.name
}

func (s *SearchView) Root() tview.Primitive {
	return s.root
}
