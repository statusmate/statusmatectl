package tui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// PromptMode distinguishes command navigation from table search.
type PromptMode int

const (
	PromptModeCommand PromptMode = iota
	PromptModeSearch
)

// commandList is the ordered list of navigable resources.
var commandList = []string{
	"incidents",
	"components",
	"maintenance",
	"team",
	"templates",
	"pages",
	"subscribers",
	"server",
	"quit",
}

// CommandPrompt is a bar that appears above the table, activated by ':' or '?'.
type CommandPrompt struct {
	*tview.TextView
	app      *App
	mode     PromptMode
	text     string
	suggest  string
	sugIdx   int
	filtered []string
	active   bool
	onFilter func(string)
	onClear  func()
}

func newCommandPrompt(app *App) *CommandPrompt {
	p := &CommandPrompt{
		TextView: tview.NewTextView(),
		app:      app,
		sugIdx:   -1,
	}
	p.SetDynamicColors(true)
	p.SetBorderPadding(0, 0, 1, 1)
	p.SetInputCapture(p.onKey)
	p.SetBackgroundColor(tcell.ColorBlack)
	return p
}

// ActivateCommand opens the prompt in command (navigation) mode.
func (p *CommandPrompt) ActivateCommand() {
	p.mode = PromptModeCommand
	p.onFilter = nil
	p.onClear = nil
	p.open(tcell.ColorWhite)
}

// ActivateSearch opens the prompt in search (filter) mode.
func (p *CommandPrompt) ActivateSearch(onFilter func(string), onClear func()) {
	p.mode = PromptModeSearch
	p.onFilter = onFilter
	p.onClear = onClear
	p.open(tcell.ColorCornflowerBlue)
}

func (p *CommandPrompt) open(borderColor tcell.Color) {
	p.text = ""
	p.suggest = ""
	p.sugIdx = -1
	p.filtered = nil
	p.active = true
	p.SetBorder(true)
	p.SetBorderColor(borderColor)
	p.updateSuggestions()
	p.draw()
	p.app.layout.ResizeItem(p, 3, 0)
	p.app.tv.SetFocus(p)
}

// Deactivate closes the prompt and returns focus to the page content.
func (p *CommandPrompt) Deactivate() {
	p.active = false
	p.text = ""
	p.suggest = ""
	p.onFilter = nil
	p.onClear = nil
	p.SetBorder(false)
	p.Clear()
	p.app.layout.ResizeItem(p, 0, 0)
	p.app.tv.SetFocus(p.app.pages)
}

func (p *CommandPrompt) onKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEscape:
		if p.mode == PromptModeSearch && p.onClear != nil {
			p.onClear()
		}
		p.Deactivate()

	case tcell.KeyBackspace2, tcell.KeyBackspace, tcell.KeyDelete:
		if len(p.text) > 0 {
			runes := []rune(p.text)
			p.text = string(runes[:len(runes)-1])
			p.updateSuggestions()
			p.draw()
			if p.mode == PromptModeSearch && p.onFilter != nil {
				p.onFilter(p.text)
			}
		}

	case tcell.KeyRune:
		if r := ev.Rune(); isValidInputRune(r) {
			p.text += string(r)
			p.updateSuggestions()
			p.draw()
			if p.mode == PromptModeSearch && p.onFilter != nil {
				p.onFilter(p.text)
			}
		}

	case tcell.KeyTab, tcell.KeyRight:
		if p.mode == PromptModeCommand && p.suggest != "" {
			p.text += p.suggest
			p.suggest = ""
			p.filtered = nil
			p.sugIdx = -1
			p.draw()
		}

	case tcell.KeyUp:
		if p.mode == PromptModeCommand {
			p.cycle(-1)
		}

	case tcell.KeyDown:
		if p.mode == PromptModeCommand {
			p.cycle(+1)
		}

	case tcell.KeyCtrlU, tcell.KeyCtrlW:
		p.text = ""
		p.updateSuggestions()
		p.draw()
		if p.mode == PromptModeSearch && p.onFilter != nil {
			p.onFilter("")
		}

	case tcell.KeyEnter:
		if p.mode == PromptModeCommand {
			p.execute()
		} else {
			p.Deactivate()
		}
	}
	return nil
}

func (p *CommandPrompt) cycle(dir int) {
	if len(p.filtered) == 0 {
		return
	}
	p.sugIdx = (p.sugIdx + dir + len(p.filtered)) % len(p.filtered)
	full := p.filtered[p.sugIdx]
	lower := strings.ToLower(p.text)
	if len(full) > len(lower) {
		p.suggest = full[len(lower):]
	} else {
		p.suggest = ""
	}
	p.draw()
}

func (p *CommandPrompt) updateSuggestions() {
	if p.mode != PromptModeCommand {
		return
	}
	p.filtered = nil
	p.sugIdx = -1
	p.suggest = ""
	lower := strings.ToLower(p.text)

	if strings.HasPrefix(lower, "server ") {
		prefix := lower[len("server "):]
		servers, _ := listAvailableServers()
		for _, s := range servers {
			if strings.HasPrefix(s, prefix) {
				p.filtered = append(p.filtered, "server "+s)
			}
		}
	} else {
		for _, s := range commandList {
			if strings.HasPrefix(s, lower) {
				p.filtered = append(p.filtered, s)
			}
		}
	}

	if len(p.filtered) > 0 {
		p.sugIdx = 0
		if len(p.filtered[0]) > len(lower) {
			p.suggest = p.filtered[0][len(lower):]
		}
	}
}

func (p *CommandPrompt) draw() {
	p.Clear()

	switch p.mode {
	case PromptModeCommand:
		if p.suggest != "" {
			_, _ = fmt.Fprintf(
				p,
				"[yellow::b]>[white::-] %s[gray::]%s[-:-:-]",
				p.text,
				p.suggest,
			)
		} else {
			_, _ = fmt.Fprintf(
				p,
				"[yellow::b]>[white::-] %s",
				p.text,
			)
		}

	case PromptModeSearch:
		_, _ = fmt.Fprintf(
			p,
			"[cornflowerblue::b]/[white::-] %s",
			p.text,
		)
	}
}

func (p *CommandPrompt) execute() {
	cmd := strings.ToLower(strings.TrimSpace(p.text + p.suggest))
	if cmd == "" && len(p.filtered) > 0 {
		cmd = p.filtered[0]
	}
	p.Deactivate()

	if strings.HasPrefix(cmd, "server") {
		rest := strings.TrimSpace(cmd[len("server"):])
		if rest != "" {
			p.app.switchServer(rest)
			return
		}
	}

	switch cmd {
	case "incidents":
		p.app.switchTo(viewIncidents)
	case "components":
		p.app.switchTo(viewComponents)
	case "maintenance":
		p.app.switchTo(viewMaintenance)
	case "team":
		p.app.switchTo(viewTeam)
	case "templates":
		p.app.switchTo(viewTemplates)
	case "server":
		p.app.switchTo(viewServers)
	case "quit":
		p.app.Quit()
	}
}

func isValidInputRune(r rune) bool {
	if unicode.IsControl(r) {
		return false
	}
	return unicode.IsPrint(r)
}
