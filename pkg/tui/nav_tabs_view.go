package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	navTabsWidth      = 20
)

// NavTabs is the navigation panel showing view shortcuts in two columns.
type NavTabs struct {
	*tview.Table
	app *App
}

func newNavTabs(app *App) *NavTabs {
	n := &NavTabs{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	n.SetBorderPadding(0, 0, 1, 0)
	return n
}

func (n *NavTabs) render() {
	items := []struct{ view, label, key string }{
		{viewIncidents, "Incidents", "i"},
		{viewComponents, "Components", "c"},
		{viewMaintenance, "Maintenance", "m"},
		{viewTemplates, "Templates", "t"},
	}

	n.Clear()
	for row, item := range items {
		keyCell := tview.NewTableCell(fmt.Sprintf("<%s>", item.key)).
			SetTextColor(tcell.ColorBlue)

		nameCell := tview.NewTableCell(" " + item.label).SetExpansion(1)
		if n.app.current == item.view {
			nameCell.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
		} else {
			nameCell.SetTextColor(tcell.ColorGray)
		}

		n.SetCell(row, 0, keyCell)
		n.SetCell(row, 1, nameCell)
	}
}
