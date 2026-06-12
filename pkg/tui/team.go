package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// TeamView displays team members.
type TeamView struct {
	app        *App
	table      *tview.Table
	users      []api.TeamUserExpanded
	displayed  []api.TeamUserExpanded
	filterText string
}

func newTeamView(app *App) *TeamView {
	v := &TeamView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))

	v.table.SetBorder(true)
	v.table.SetTitle(" Team ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetBackgroundColor(tcell.ColorBlack)

	return v
}

func (v *TeamView) root() tview.Primitive { return v.table }

func (v *TeamView) refresh() {
	go func() {
		result, err := v.app.client.GetPaginatedTeamUsersExpanded(
			api.NewAllPaginatedRequest(api.PaginatedRequestFilter{}),
		)
		if err != nil {
			return
		}
		v.users = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *TeamView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *TeamView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *TeamView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, u := range v.users {
		if lower == "" ||
			strings.Contains(strings.ToLower(u.User.Username), lower) ||
			strings.Contains(strings.ToLower(u.User.Email), lower) ||
			strings.Contains(strings.ToLower(u.Role), lower) {
			v.displayed = append(v.displayed, u)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Team [%d/%d] ", len(v.displayed), len(v.users)))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Team [%d] ", len(v.users)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "ROLE", "USERNAME", "EMAIL", "ACTIVE"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, u := range v.displayed {
		row := i + 1
		active := "yes"
		if !u.IsActive {
			active = "no"
		}
		v.table.SetCell(row, 0, tview.NewTableCell(shortUUID(u.UUID)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(u.Role).SetTextColor(tcell.ColorAqua).SetExpansion(1))
		v.table.SetCell(row, 2, tview.NewTableCell(u.User.Username).SetTextColor(tcell.ColorWhite).SetExpansion(2))
		v.table.SetCell(row, 3, tview.NewTableCell(u.User.Email).SetTextColor(tcell.ColorGray).SetExpansion(3))
		v.table.SetCell(row, 4, tview.NewTableCell(active).SetTextColor(tcell.ColorGray))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}
