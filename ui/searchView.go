package ui

import (
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
)

type SearchView struct {
	App     *App
	name    string
	root    *tview.Grid
	title   *tview.TextView
	input   *tview.InputField
	results *tview.List
	active  bool
}

func NewSearchView(app *App) *SearchView {
	searchView := &SearchView{
		App:     app,
		name:    "SearchView",
		root:    tview.NewGrid(),
		title:   tview.NewTextView(),
		input:   tview.NewInputField(),
		results: tview.NewList(),
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

func (s *SearchView) SetupEvents() {
	s.input.SetDoneFunc(func(key tcell.Key) {
		s.results.Clear()

		var urls []string
		for _, el := range s.App.urls {
			urls = append(urls, el.Url)
		}

		results := fuzzy.RankFind(s.input.GetText(), urls)
		sort.Sort(results)

		for _, item := range results {
			s.results.AddItem(item.Target, "", 0, nil)
		}
		s.App.SetFocus(s.results)
	})

	s.results.SetSelectedFunc(func(_ int, result string, _ string, _ rune) {
		if s.results.GetItemCount() == 0 {
			return
		}
		mainview := s.App.views["MainView"].(*MainView)
		idxs := mainview.urlsList.FindItems(result, "", false, true)
		mainview.urlsList.SetCurrentItem(idxs[0])
		s.App.SwitchToPage("MainView")
	})

	s.results.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			s.App.SetFocus(s.input)
			return nil
		} else if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		} else if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		} else if event.Rune() == 'q' {
			s.App.SwitchToPage("MainView")
		} else if event.Rune() == '/' {
			s.App.SetFocus(s.input)
		}
		return event
	})
}
