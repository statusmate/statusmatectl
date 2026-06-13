package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// PublicPagesView displays all available status pages.
type PublicPagesView struct {
	app        *App
	table      *tview.Table
	describe   *PublicPageDescribeView
	pages      []api.StatusPage
	displayed  []api.StatusPage
	filterText string
}

func newPublicPagesView(app *App) *PublicPagesView {
	v := &PublicPagesView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetTitle(" Status Pages ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetBackgroundColor(tcell.ColorBlack)
	v.table.SetInputCapture(v.onKey)

	v.describe = newPublicPageDescribeView(app)

	return v
}

func (v *PublicPagesView) root() tview.Primitive { return v.table }

func (v *PublicPagesView) refresh() {
	go func() {
		result, err := v.app.client.GetPaginatedStatusPages(api.NewAllPaginatedRequest(nil))
		if err != nil {
			return
		}
		v.pages = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *PublicPagesView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *PublicPagesView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *PublicPagesView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, p := range v.pages {
		if lower == "" ||
			strings.Contains(strings.ToLower(p.Name), lower) ||
			strings.Contains(strings.ToLower(p.Slug), lower) ||
			strings.Contains(strings.ToLower(p.AbsoluteURL), lower) ||
			strings.Contains(strings.ToLower(p.TeamSlug), lower) {
			v.displayed = append(v.displayed, p)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Status Pages [%d/%d] [green::]</%s>[-:-:-] ", len(v.displayed), len(v.pages), lower))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Status Pages [%d] ", len(v.pages)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "NAME", "SLUG", "IMPACT", "TEAM"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, p := range v.displayed {
		row := i + 1
		v.table.SetCell(row, 0, tview.NewTableCell(shortUUID(p.UUID)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(p.Name).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 2, tview.NewTableCell(p.Slug).SetTextColor(tcell.ColorCornflowerBlue).SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(string(p.Impact)).SetTextColor(impactColor(p.Impact)).SetExpansion(2))
		v.table.SetCell(row, 4, tview.NewTableCell(p.TeamSlug).SetTextColor(tcell.ColorGray).SetExpansion(2))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *PublicPagesView) selected() *api.StatusPage {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *PublicPagesView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEnter:
		if p := v.selected(); p != nil {
			v.app.setCurrentPage(p)
		}
		return nil
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'd':
			if p := v.selected(); p != nil {
				v.describe.show(p)
			}
			return nil
		}
	}
	return ev
}
