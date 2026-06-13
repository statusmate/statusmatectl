package tui

import (
	"os/exec"
	"runtime"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// PublicPageDescribeView shows the detail of a single status page.
type PublicPageDescribeView struct {
	app          *App
	detail       *tview.Table
	detailURLRow int
	detailURL    string
}

func newPublicPageDescribeView(app *App) *PublicPageDescribeView {
	d := &PublicPageDescribeView{app: app}

	d.detail = tview.NewTable().SetSelectable(true, false)
	d.detail.SetBorder(true)
	d.detail.SetTitle(" Page Detail ")
	d.detail.SetTitleAlign(tview.AlignCenter)
	d.detail.SetBackgroundColor(tcell.ColorBlack)
	d.detail.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEscape:
			app.popPage()
			app.tv.SetFocus(app.publicPages.table)
			return nil
		case tcell.KeyEnter:
			row, _ := d.detail.GetSelection()
			if row == d.detailURLRow && d.detailURL != "" {
				openBrowser(d.detailURL)
			}
			return nil
		}
		return ev
	})
	app.pages.AddPage(viewPubPageDescribe, d.detail, true, false)

	return d
}

func (d *PublicPageDescribeView) show(p *api.StatusPage) {
	d.detail.Clear()

	row := 0
	d.detail.SetCell(row, 0, detailLabelCell("Name"))
	d.detail.SetCell(row, 1, detailValueCell(p.Name))
	row++
	d.detail.SetCell(row, 0, detailLabelCell("UUID"))
	d.detail.SetCell(row, 1, detailValueCell(p.UUID))
	row++
	d.detail.SetCell(row, 0, detailLabelCell("Slug"))
	d.detail.SetCell(row, 1, detailValueCell(p.Slug))
	row++
	d.detailURLRow = row
	d.detailURL = p.AbsoluteURL
	urlCell := tview.NewTableCell(p.AbsoluteURL).
		SetTextColor(tcell.ColorCornflowerBlue).
		SetAttributes(tcell.AttrUnderline).
		SetExpansion(1)
	d.detail.SetCell(row, 0, detailLabelCell("URL"))
	d.detail.SetCell(row, 1, urlCell)
	row++
	d.detail.SetCell(row, 0, detailLabelCell("Team"))
	d.detail.SetCell(row, 1, detailValueCell(p.TeamSlug))
	row++
	d.detail.SetCell(row, 0, detailLabelCell("Impact"))
	d.detail.SetCell(row, 1, tview.NewTableCell(string(p.Impact)).
		SetTextColor(impactColor(p.Impact)).SetExpansion(1))
	row++
	if p.CustomDomain != "" {
		d.detail.SetCell(row, 0, detailLabelCell("Custom domain"))
		d.detail.SetCell(row, 1, detailValueCell(p.CustomDomain))
		row++
	}
	d.detail.SetCell(row, 0, detailLabelCell("Timezone"))
	d.detail.SetCell(row, 1, detailValueCell(p.TimeZone))
	row++
	d.detail.SetCell(row, 0, detailLabelCell("Language"))
	d.detail.SetCell(row, 1, detailValueCell(string(p.Lang)))
	row++
	d.detail.SetCell(row, 0, detailLabelCell("Dark mode"))
	d.detail.SetCell(row, 1, detailValueCell(string(p.DarkMode)))
	row++
	if p.CreatedAt != nil {
		d.detail.SetCell(row, 0, detailLabelCell("Created"))
		d.detail.SetCell(row, 1, detailValueCell(formatTimePtr(p.CreatedAt)))
		row++
	}

	d.app.pushPage(viewPubPageDescribe)
	d.app.tv.SetFocus(d.detail)
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "start"
	default:
		cmd = "xdg-open"
	}
	exec.Command(cmd, url).Start() //nolint:errcheck
}
