package tui

import (
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const viewSplash = "splash"

type SplashView struct {
	app  *App
	text *tview.TextView
}

func newSplashView(app *App) *SplashView {
	s := &SplashView{app: app}

	s.text = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetScrollable(true).
		SetWrap(false)
	s.text.SetBackgroundColor(tcell.ColorBlack)
	s.text.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		app.dismiss()
		return nil
	})

	app.pages.AddPage(viewSplash, s.text, true, false)
	return s
}

func (s *SplashView) show() {
	s.text.SetText(splashContent())
	s.app.pages.Reset(viewSplash)
	s.app.tv.SetFocus(s.text)
	s.app.layout.ResizeItem(s.app.header, 0, 0)
	s.app.layout.ResizeItem(s.app.breadcrumbs, 0, 0)

	go func() {
		time.Sleep(2 * time.Second)
		s.app.tv.QueueUpdateDraw(s.app.dismiss)
	}()
}

func splashContent() string {
	return `


[white::]  ____________________________________
 /                                    \
 |   [cyan::b]  ___ _        _                  [-:-:-][white::]  |
 |   [cyan::b] / __| |_ __ _| |_ _  _ ___       [-:-:-][white::]  |
 |   [cyan::b] \__ \  _/ _  |  _| || (_-<       [-:-:-][white::]  |
 |   [cyan::b] |___/\__\__,_|\__|\_,_/__/       [-:-:-][white::]  |
 |   [cyan::b]        _ __  __ _  __ _| |_ ___  [-:-:-][white::]  |
 |   [cyan::b]       | '  \/ _  |/ _  |  _/ -_) [-:-:-][white::]  |
 |   [cyan::b]       |_|_|_\__,_|\__,_|\__\___| [-:-:-][white::]  |
 |                                       |
 |   [#606060::]Press any key to continue...     [-:-:-][white::]   |
 \____________________________________/
[-:-:-][white::]

        [yellow::b]^__^[-:-:-]
                [yellow::b](oo)[-:-:-][white::]\_______[-:-:-]
[white::]                   (__)\       )\/\[-:-:-]
[white::]                       ||----w |[-:-:-]
[white::]                       ||     ||[-:-:-]
`
}
