package tui

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const rowsPerColumn = 5

// PageActions shows page-specific subcommands and shortcuts in the header.
// Content changes depending on the current active view.
type PageActions struct {
	*tview.Table
	app *App
}

func newPageActions(app *App) *PageActions {
	p := &PageActions{
		Table: tview.NewTable().SetSelectable(false, false),
		app:   app,
	}
	p.SetBorderPadding(0, 0, 1, 0)
	return p
}

func (p *PageActions) render() {
	p.Clear()
	switch p.app.current {
	case viewRequestLogs:
		p.renderLogs()
	case viewServers:
		p.renderServers()
	case viewIncidents:
		p.renderIncidents()
	case viewComponents:
		p.renderComponentsMaintenance()
	case viewMaintenance:
		p.renderMaintenance()
	case viewTeam:
		p.renderGlobal(0)
	case viewTemplates:
		p.renderTemplates()
	case viewPublicPages:
		p.renderPublicPages()
	case viewSubscribers:
		p.renderSubscribers()
	case viewMaintDescribe:
		p.renderMaintDescribe()
	}
}

// renderLogs shows time-filter presets (3 rows × 2 cols) plus global shortcuts.
func (p *PageActions) renderLogs() {
	activePreset := p.app.logs.preset
	// Preset pairs per row: (left index, right index)
	pairs := [][2]int{{0, 3}, {1, 4}, {2, 5}}
	for row, pair := range pairs {
		for col, idx := range pair {
			preset := logPresetList[idx]
			p.setPresetCell(row, col*2, preset.key, preset.label, idx == activePreset)
		}
	}
	p.setGlobalRow(3, ":", "Cmd", "?", "Search")
	p.setGlobalRow(4, "r", "Refresh", "q", "Quit")
}

func (p *PageActions) renderMaintDescribe() {
	p.renderGlobal(0)
}

func (p *PageActions) renderServers() {
	p.setPageAction(0, "d", "Describe")
	p.setPageAction(1, "l", "Logs")
	p.renderGlobal(2)
}

func (p *PageActions) renderIncidents() {
	p.setPageAction(0, "enter", "Describe")
	p.setPageAction(1, "n", "New")
	p.setPageAction(2, "u", "Update")
	p.setPageAction(3, "d", "Delete")
	p.setPageAction(4, "shift+r", "Resolve")
	p.setPageAction(5, "o", "Open in browser")
	p.renderGlobal(6)
}

func (p *PageActions) renderComponentsMaintenance() {
	p.setPageAction(0, "enter", "Describe")
	p.setPageAction(1, "l", "Log")
	p.renderGlobal(2)
}

func (p *PageActions) renderMaintenance() {
	p.setPageAction(0, "enter", "Describe")
	p.setPageAction(1, "d", "Delete")
	p.setPageAction(2, "o", "Open in browser")
	p.renderGlobal(3)
}

func (p *PageActions) renderTemplates() {
	p.setPageAction(0, "enter", "Describe")
	p.setPageAction(1, "n", "New from template")
	p.setPageAction(2, "d", "Delete")
	p.renderGlobal(3)
}

func (p *PageActions) renderPublicPages() {
	p.setPageAction(0, "enter", "Set current")
	p.setPageAction(1, "d", "Describe")
	p.renderGlobal(2)
}

func (p *PageActions) renderSubscribers() {
	p.setPageAction(0, "enter", "Describe")
	p.setPageAction(1, "n", "New")
	p.setPageAction(2, "a", "Approve")
	p.setPageAction(3, "d", "Delete")
	p.renderGlobal(4)
}

// renderGlobal appends global shortcuts starting at the given row.
func (p *PageActions) renderGlobal(startRow int) {
	globals := [][2]string{
		{":", "Command"},
		{"?", "Search"},
		{"r", "Refresh"},
		{"q", "Quit"},
	}
	for i, g := range globals {
		p.setPageAction(startRow+i, g[0], g[1])
	}
}

// setPageAction renders a single page-specific action key+label at the given row.
func (p *PageActions) setPageAction(row int, key, label string) {
	col := row / rowsPerColumn // 0 → колонки 0/1, 1 → 2/3, ...
	r := row % rowsPerColumn
	keyCol, labelCol := col*2, col*2+1

	p.SetCell(r, keyCol, tview.NewTableCell(fmt.Sprintf("<%s>", key)).
		SetTextColor(tcell.ColorBlue).SetSelectable(false))
	p.SetCell(r, labelCol, tview.NewTableCell(" "+label).
		SetExpansion(1).SetTextColor(tcell.ColorGray).SetSelectable(false))
}

// setPresetCell renders a log time-preset key+label.
// col*2 and col*2+1 are used so two presets fit per row.
func (p *PageActions) setPresetCell(row, colBase int, key, label string, active bool) {
	p.SetCell(row, colBase, tview.NewTableCell(fmt.Sprintf("<%s>", key)).
		SetTextColor(tcell.ColorMediumPurple).SetSelectable(false))

	labelText := " " + label + "  "
	lc := tview.NewTableCell(labelText).SetSelectable(false)
	if active {
		lc.SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold)
	} else {
		lc.SetTextColor(tcell.ColorWhite)
	}
	p.SetCell(row, colBase+1, lc)
}

// handleKey processes page-specific key events before the global handler.
// Returns nil if the event was consumed, or the original event to pass through.
func (p *PageActions) handleKey(ev *tcell.EventKey) *tcell.EventKey {
	switch p.app.current {
	case viewRequestLogs:
		r := ev.Rune()
		if r >= '0' && r <= '5' {
			p.app.logs.setPreset(int(r - '0'))
			return nil
		}

	case viewServers:
		switch ev.Rune() {
		case 'l':
			p.app.switchTo(viewRequestLogs)
		}

	}
	return ev
}

// setGlobalRow renders two global shortcuts side-by-side on the given row.
func (p *PageActions) setGlobalRow(row int, key1, label1, key2, label2 string) {
	p.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("<%s>", key1)).
		SetTextColor(tcell.ColorYellow).SetSelectable(false))
	p.SetCell(row, 1, tview.NewTableCell(" "+label1+"  ").
		SetTextColor(tcell.ColorWhite).SetSelectable(false))
	p.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("<%s>", key2)).
		SetTextColor(tcell.ColorYellow).SetSelectable(false))
	p.SetCell(row, 3, tview.NewTableCell(" "+label2).
		SetExpansion(1).SetTextColor(tcell.ColorWhite).SetSelectable(false))
}
