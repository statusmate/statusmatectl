package tui

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/statusmate/statusmatectl/pkg/api"
)

// SubscribersView lists the subscribers of the current status page and handles
// creating, approving and deleting them.
type SubscribersView struct {
	app          *App
	table        *tview.Table
	describe     *SubscriberDescribeView
	deleteModal  *tview.Modal
	approveModal *tview.Modal
	subscribers  []api.Subscriber
	displayed    []api.Subscriber
	filterText   string
}

func newSubscribersView(app *App) *SubscribersView {
	v := &SubscribersView{app: app}

	v.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.StyleDefault.
			Background(tcell.ColorNavy).
			Foreground(tcell.ColorWhite))

	v.table.SetBorder(true)
	v.table.SetTitle(" Subscribers ")
	v.table.SetTitleAlign(tview.AlignCenter)
	v.table.SetBackgroundColor(tcell.ColorBlack)
	v.table.SetInputCapture(v.onKey)

	v.deleteModal = tview.NewModal().
		SetText("Delete this subscriber?").
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(_ int, label string) {
			app.popPage()
			app.tv.SetFocus(v.table)
			if label == "Delete" {
				v.confirmDelete()
			}
		})
	app.pages.AddPage("subDelete", v.deleteModal, true, false)

	v.approveModal = tview.NewModal().
		SetText("Approve this subscriber?").
		AddButtons([]string{"Approve", "Cancel"}).
		SetDoneFunc(func(_ int, label string) {
			app.popPage()
			app.tv.SetFocus(v.table)
			if label == "Approve" {
				v.confirmApprove()
			}
		})
	app.pages.AddPage("subApprove", v.approveModal, true, false)

	v.describe = newSubscriberDescribeView(app)

	return v
}

func (v *SubscribersView) root() tview.Primitive { return v.table }

func (v *SubscribersView) refresh() {
	if v.app.statusPage == nil {
		return
	}
	statusPageID := v.app.statusPage.ID
	go func() {
		result, err := v.app.client.GetPaginatedSubscribers(
			api.NewAllPaginatedRequest(api.PaginatedRequestFilter{"status_page": statusPageID}),
		)
		if err != nil {
			return
		}
		v.subscribers = result.Results
		v.app.tv.QueueUpdateDraw(v.render)
	}()
}

func (v *SubscribersView) filter(text string) {
	v.filterText = text
	v.render()
}

func (v *SubscribersView) clearFilter() {
	v.filterText = ""
	v.render()
}

func (v *SubscribersView) render() {
	lower := strings.ToLower(v.filterText)
	v.displayed = v.displayed[:0]
	for _, s := range v.subscribers {
		if lower == "" || strings.Contains(strings.ToLower(s.Email), lower) {
			v.displayed = append(v.displayed, s)
		}
	}

	if lower != "" {
		v.table.SetTitle(fmt.Sprintf(" Subscribers [%d/%d] [green::]</%s>[-:-:-] ", len(v.displayed), len(v.subscribers), lower))
	} else {
		v.table.SetTitle(fmt.Sprintf(" Subscribers [%d] ", len(v.subscribers)))
	}
	v.table.Clear()

	for i, h := range []string{"UUID", "EMAIL", "BY EMAIL", "BY WEBHOOK", "VERIFIED", "CREATED"} {
		v.table.SetCell(0, i, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false).
			SetExpansion(1))
	}

	for i, s := range v.displayed {
		row := i + 1
		v.table.SetCell(row, 0, tview.NewTableCell(shortUUID(nullOrDash(s.UUID))).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 1, tview.NewTableCell(s.Email).SetTextColor(tcell.ColorWhite).SetExpansion(3))
		v.table.SetCell(row, 2, tview.NewTableCell(yesNo(s.SubscribeByEmail)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 3, tview.NewTableCell(yesNo(s.SubscribeByWebhook)).SetTextColor(tcell.ColorGray))
		v.table.SetCell(row, 4, tview.NewTableCell(yesNo(s.Confirmed)).SetTextColor(verifiedColor(s.Confirmed)))
		v.table.SetCell(row, 5, tview.NewTableCell(formatTimePtr(s.CreatedAt)).SetTextColor(tcell.ColorGray))
	}

	if len(v.displayed) > 0 {
		v.table.Select(1, 0)
	}
}

func (v *SubscribersView) selected() *api.Subscriber {
	row, _ := v.table.GetSelection()
	if row <= 0 || row-1 >= len(v.displayed) {
		return nil
	}
	return &v.displayed[row-1]
}

func (v *SubscribersView) onKey(ev *tcell.EventKey) *tcell.EventKey {
	switch ev.Key() {
	case tcell.KeyEnter:
		if s := v.selected(); s != nil {
			v.describe.show(s)
		}
		return nil
	}
	switch ev.Rune() {
	case 'n':
		v.showCreateForm()
		return nil
	case 'a':
		if s := v.selected(); s != nil && !s.Confirmed {
			v.showApproveConfirm(s)
		}
		return nil
	case 'd':
		if s := v.selected(); s != nil {
			v.showDeleteConfirm(s)
		}
		return nil
	}
	return ev
}

func (v *SubscribersView) showDeleteConfirm(s *api.Subscriber) {
	v.deleteModal.SetText(fmt.Sprintf("Delete subscriber:\n[white::b]%s[-:-:-]?", s.Email))
	v.app.pushPage("subDelete")
	v.app.tv.SetFocus(v.deleteModal)
}

func (v *SubscribersView) confirmDelete() {
	s := v.selected()
	if s == nil || s.UUID == nil {
		return
	}
	uuid := *s.UUID
	go func() {
		if err := v.app.client.DeleteSubscriber(uuid); err != nil {
			return
		}
		v.app.tv.QueueUpdateDraw(v.refresh)
	}()
}

func (v *SubscribersView) showApproveConfirm(s *api.Subscriber) {
	v.approveModal.SetText(fmt.Sprintf("Approve subscriber:\n[white::b]%s[-:-:-]?", s.Email))
	v.app.pushPage("subApprove")
	v.app.tv.SetFocus(v.approveModal)
}

func (v *SubscribersView) confirmApprove() {
	s := v.selected()
	if s == nil || s.UUID == nil {
		return
	}
	uuid := *s.UUID
	go func() {
		if err := v.app.client.VerifySubscriber(uuid); err != nil {
			return
		}
		v.app.tv.QueueUpdateDraw(v.refresh)
	}()
}

func (v *SubscribersView) showCreateForm() {
	if v.app.statusPage == nil {
		return
	}
	statusPageID := v.app.statusPage.ID

	const pageName = "subCreate"
	close := func() {
		v.app.pages.RemovePage(pageName)
		v.app.tv.SetFocus(v.table)
	}

	form := tview.NewForm()
	form.AddInputField("Email", "", 40, nil, nil)
	form.AddButton("Create", func() {
		email := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		close()
		if email == "" {
			return
		}
		go func() {
			payload := &api.CreateSubscriberPayload{
				Email:            email,
				StatusPage:       statusPageID,
				SubscribeByEmail: true,
			}
			if _, err := v.app.client.CreateSubscriber(payload); err != nil {
				return
			}
			v.app.tv.QueueUpdateDraw(v.refresh)
		}()
	})
	form.AddButton("Cancel", close)
	form.SetBorder(true)
	form.SetTitle(" New Subscriber ")
	form.SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape {
			close()
			return nil
		}
		return ev
	})

	v.app.pages.AddPage(pageName, modalWrap(form, 50, 7), true, true)
	v.app.tv.SetFocus(form)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func verifiedColor(verified bool) tcell.Color {
	if verified {
		return tcell.ColorGreen
	}
	return tcell.ColorOrange
}

func nullOrDash(s *string) string {
	if s == nil {
		return "-"
	}
	return *s
}

// modalWrap centers a primitive at the given width/height, like a modal.
func modalWrap(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}
