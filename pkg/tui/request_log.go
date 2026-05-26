package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

type httpLogEntry struct {
	ts       time.Time
	method   string
	url      string
	status   string
	request  string
	response string
}

type logPresetInfo struct {
	key   string
	label string
	dur   time.Duration // 0 = no time filter (tail); negative = head; positive = last N
}

var logPresetList = []logPresetInfo{
	{"0", "tail", 0},
	{"1", "head", -1},
	{"2", "1m", time.Minute},
	{"3", "5m", 5 * time.Minute},
	{"4", "15m", 15 * time.Minute},
	{"5", "30m", 30 * time.Minute},
}

// RequestLogView displays HTTP request logs from the current server's log file.
type RequestLogView struct {
	app        *App
	table      *tview.Table
	detail     *tview.TextView
	entries    []httpLogEntry // all entries, chronological order (oldest first)
	displayed  []httpLogEntry
	filterText string
	preset     int // index into logPresetList
	mu         sync.Mutex
	lastOffset int64
	stopCh     chan struct{}
}

func newRequestLogView(app *App) *RequestLogView {
	v := &RequestLogView{app: app, preset: 0}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))
	v.table.SetBorder(true)
	v.table.SetTitle(" HTTP Logs ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetInputCapture(v.onKey)

	v.detail = tview.NewTextView()
	v.detail.SetBorder(true)
	v.detail.SetTitle(" Request Detail ")
	v.detail.SetTitleAlign(tview.AlignCenter)
	v.detail.SetDynamicColors(true)
	v.detail.SetScrollable(true)
	v.detail.SetWrap(false)
	v.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			app.pages.SwitchToPage(requestViewLogs)
			app.tv.SetFocus(v.table)
		}
		return ev
	})
	app.pages.AddPage("logDetail", v.detail, true, false)

	return v
}

func (v *RequestLogView) root() tview.Primitive { return v.table }

func (v *RequestLogView) logFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	domain := sanitizeDomain(strings.TrimPrefix(
		strings.TrimPrefix(v.app.client.BaseURL, "https://"), "http://"))
	return filepath.Join(homeDir, st4Dir, domain, "http_requests.log")
}

func (v *RequestLogView) refresh() {
	v.stopTailing()
	go func() {
		entries, offset := v.readFullFile()
		v.mu.Lock()
		v.lastOffset = offset
		v.mu.Unlock()
		v.app.tv.QueueUpdateDraw(func() {
			v.entries = entries
			v.render()
		})
		v.startTailing()
	}()
}

func (v *RequestLogView) readFullFile() ([]httpLogEntry, int64) {
	f, err := os.Open(v.logFilePath())
	if err != nil {
		return nil, 0
	}
	defer f.Close()

	var entries []httpLogEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		if e := parseHTTPLogLine(scanner.Text()); e != nil {
			entries = append(entries, *e)
		}
	}

	offset, _ := f.Seek(0, io.SeekCurrent)
	return entries, offset
}

func (v *RequestLogView) startTailing() {
	v.mu.Lock()
	if v.stopCh != nil {
		v.mu.Unlock()
		return
	}
	ch := make(chan struct{})
	v.stopCh = ch
	v.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ch:
				return
			case <-ticker.C:
				v.tailOnce()
			}
		}
	}()
}

func (v *RequestLogView) stopTailing() {
	v.mu.Lock()
	ch := v.stopCh
	v.stopCh = nil
	v.mu.Unlock()
	if ch != nil {
		close(ch)
	}
}

func (v *RequestLogView) tailOnce() {
	path := v.logFilePath()
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return
	}

	v.mu.Lock()
	offset := v.lastOffset
	v.mu.Unlock()

	// File was truncated or rotated — re-read from start
	if info.Size() < offset {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return
		}
		var entries []httpLogEntry
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
		for scanner.Scan() {
			if e := parseHTTPLogLine(scanner.Text()); e != nil {
				entries = append(entries, *e)
			}
		}
		newOffset, _ := f.Seek(0, io.SeekCurrent)
		v.mu.Lock()
		v.lastOffset = newOffset
		v.mu.Unlock()
		v.app.tv.QueueUpdateDraw(func() {
			v.entries = entries
			v.render()
		})
		return
	}

	if info.Size() == offset {
		return
	}

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return
	}

	var newEntries []httpLogEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		if e := parseHTTPLogLine(scanner.Text()); e != nil {
			newEntries = append(newEntries, *e)
		}
	}

	newOffset, _ := f.Seek(0, io.SeekCurrent)
	v.mu.Lock()
	v.lastOffset = newOffset
	v.mu.Unlock()

	if len(newEntries) == 0 {
		return
	}

	v.app.tv.QueueUpdateDraw(func() {
		v.entries = append(v.entries, newEntries...)
		v.render()
	})
}

func (v *RequestLogView) setPreset(idx int) {
	v.preset = idx
	v.app.renderHeader()
	v.render()
}

func (v *RequestLogView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *RequestLogView) clearFilter() {
	v.filterText = ""
	v.render()
}


func (v *RequestLogView) render() {
	lower := strings.ToLower(v.filterText)

	preset := logPresetList[v.preset]
	var timeFiltered []httpLogEntry
	if preset.dur > 0 {
		cutoff := time.Now().Add(-preset.dur)
		for _, e := range v.entries {
			if e.ts.After(cutoff) {
				timeFiltered = append(timeFiltered, e)
			}
		}
	} else {
		timeFiltered = v.entries
	}

	v.displayed = v.displayed[:0]
	for _, e := range timeFiltered {
		if lower == "" ||
			strings.Contains(strings.ToLower(e.method), lower) ||
			strings.Contains(strings.ToLower(e.url), lower) ||
			strings.Contains(strings.ToLower(e.status), lower) {
			v.displayed = append(v.displayed, e)
		}
	}

	domain := sanitizeDomain(strings.TrimPrefix(
		strings.TrimPrefix(v.app.client.BaseURL, "https://"), "http://"))
	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" HTTP Logs: %s [%d/%d] ", domain, len(v.displayed), len(v.entries)))
	} else {
		v.table.SetTitle(fmt.Sprintf(" HTTP Logs: %s [%d] ", domain, len(v.entries)))
	}
	v.table.Clear()

	for i, h := range []string{"TIME", "METHOD", "URL", "STATUS"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, e := range v.displayed {
		row := i + 1
		timeStr := "-"
		if !e.ts.IsZero() {
			timeStr = e.ts.Local().Format("2006-01-02 15:04:05")
		}
		statusColor := httpStatusColor(e.status)

		v.table.SetCell(row, 0, tview.NewTableCell(timeStr).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(e.method).SetTextColor(httpMethodColor(e.method)))
		v.table.SetCell(row, 2, tview.NewTableCell(e.url).SetTextColor(tcell.ColorWhite).SetExpansion(5))
		v.table.SetCell(row, 3, tview.NewTableCell(e.status).SetTextColor(statusColor))
	}

	n := len(v.displayed)
	if n > 0 {
		if v.preset == 1 { // head — scroll to top
			v.table.Select(1, 0)
		} else { // tail and time presets — autoscroll to bottom
			v.table.Select(n, 0)
		}
	}
}

func (v *RequestLogView) selected() *httpLogEntry {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *RequestLogView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	if ev.Key() == tcell.KeyEnter {
		if e := v.selected(); e != nil {
			v.showDetail(e)
		}
		return nil
	}
	r := ev.Rune()
	if r >= '0' && r <= '5' {
		v.setPreset(int(r - '0'))
		return nil
	}
	return ev
}

func (v *RequestLogView) showDetail(e *httpLogEntry) {
	v.detail.Clear()
	v.detail.SetTitle(fmt.Sprintf(" %s %s ", e.method, truncate(e.url, 60)))

	fmt.Fprintf(
		v.detail,
		"[yellow::b]%s %s[-:-:-]  [gray::]%s[-:-:-]\n\n",
		e.method,
		e.url,
		e.ts.Local().Format("2006-01-02 15:04:05"),
	)

	statusColor := "green"
	if strings.HasPrefix(e.status, "4") || strings.HasPrefix(e.status, "5") {
		statusColor = "red"
	} else if strings.HasPrefix(e.status, "3") {
		statusColor = "yellow"
	}

	fmt.Fprintf(
		v.detail,
		"[white::b]Status:[-:-:-] [%s::]%s[-:-:-]\n\n",
		statusColor,
		e.status,
	)

	if e.request != "" {
		fmt.Fprintf(
			v.detail,
			"[yellow::b]── REQUEST ──[-:-:-]\n[white::]%s[-:-:-]\n\n",
			tview.Escape(e.request),
		)
	}

	if e.response != "" {
		fmt.Fprintf(
			v.detail,
			"[yellow::b]── RESPONSE ──[-:-:-]\n[white::]%s[-:-:-]\n",
			tview.Escape(e.response),
		)
	}

	v.detail.ScrollToBeginning()
	v.app.pages.SwitchToPage("logDetail")
	v.app.tv.SetFocus(v.detail)
}

func httpMethodColor(method string) tcell.Color {
	switch method {
	case "GET":
		return tcell.ColorCornflowerBlue
	case "POST":
		return tcell.ColorGreen
	case "PATCH", "PUT":
		return tcell.ColorYellow
	case "DELETE":
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}

func httpStatusColor(status string) tcell.Color {
	if strings.HasPrefix(status, "2") {
		return tcell.ColorGreen
	}
	if strings.HasPrefix(status, "3") {
		return tcell.ColorYellow
	}
	if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
		return tcell.ColorRed
	}
	return tcell.ColorGray
}

// parseHTTPLogLine parses a slog TextHandler log line and extracts HTTP info.
func parseHTTPLogLine(line string) *httpLogEntry {
	fields := parseSlogLine(line)
	if fields["msg"] != "HTTP request" {
		return nil
	}

	e := &httpLogEntry{}

	if t := fields["time"]; t != "" {
		e.ts, _ = time.Parse(time.RFC3339Nano, t)
		if e.ts.IsZero() {
			e.ts, _ = time.Parse("2006-01-02T15:04:05.999Z07:00", t)
		}
	}

	e.url = fields["url"]

	if req := fields["request"]; req != "" {
		e.request = req
		if idx := strings.IndexByte(req, ' '); idx > 0 {
			e.method = strings.ToUpper(req[:idx])
		}
	}

	if resp := fields["response"]; resp != "" {
		e.response = resp
		if strings.HasPrefix(resp, "HTTP/") {
			statusLine := resp
			if idx := strings.Index(resp, "\r\n"); idx >= 0 {
				statusLine = resp[:idx]
			}
			parts := strings.Fields(statusLine)
			if len(parts) >= 3 {
				e.status = parts[1] + " " + parts[2]
			} else if len(parts) == 2 {
				e.status = parts[1]
			}
		}
	}

	return e
}

// parseSlogLine parses a slog TextHandler format line into key=value pairs.
func parseSlogLine(line string) map[string]string {
	result := make(map[string]string)
	i := 0
	for i < len(line) {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			break
		}
		eq := strings.Index(line[i:], "=")
		if eq < 0 {
			break
		}
		key := line[i : i+eq]
		i += eq + 1
		if i >= len(line) {
			result[key] = ""
			break
		}
		var value string
		if line[i] == '"' {
			j := i + 1
			for j < len(line) {
				if line[j] == '\\' {
					j += 2
				} else if line[j] == '"' {
					j++
					break
				} else {
					j++
				}
			}
			quoted := line[i:j]
			if unquoted, err := strconv.Unquote(quoted); err == nil {
				value = unquoted
			} else {
				value = strings.Trim(quoted, `"`)
			}
			i = j
		} else {
			j := i
			for j < len(line) && line[j] != ' ' {
				j++
			}
			value = line[i:j]
			i = j
		}
		result[key] = value
	}
	return result
}
