package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

const (
	pageSwitcherWidth = 26
)

// PageSwitcher shows numbered status pages (max 5) for quick navigation via digit keys.
type PageSwitcher struct {
	*tview.Table
	app   *App
	pages []api.StatusPage
}

func newPageSwitcher(app *App) *PageSwitcher {
	p := &PageSwitcher{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	p.SetBorderPadding(0, 0, 1, 0)
	return p
}

func (p *PageSwitcher) setPages(pages []api.StatusPage) {
	p.pages = pages
	p.render()
}

func (p *PageSwitcher) render() {
	p.Clear()
	limit := len(p.pages)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		pg := p.pages[i]

		keyCell := tview.NewTableCell(fmt.Sprintf("<%d>", i)).
			SetTextColor(tcell.ColorFuchsia)

		nameCell := tview.NewTableCell(" " + pg.Name).SetExpansion(1)
		if p.app.statusPage != nil && p.app.statusPage.ID == pg.ID {
			nameCell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
		} else {
			nameCell.SetTextColor(tcell.ColorGray)
		}

		p.SetCell(i, 0, keyCell)
		p.SetCell(i, 1, nameCell)
	}
}
