package tui

import "github.com/derailed/tview"

// Pages wraps tview.Pages with a navigation Stack.
// Stack listeners are notified on push/pop/reset; the Pages itself
// listens to StackTop and switches the visible tview page accordingly.
type Pages struct {
	*tview.Pages
	*Stack
}

func newPages() *Pages {
	p := &Pages{
		Pages: tview.NewPages(),
		Stack: newStack(),
	}
	p.Stack.addListener(p)
	return p
}

// StackPushed implements StackListener.
func (p *Pages) StackPushed(_ string) {}

// StackPopped implements StackListener.
func (p *Pages) StackPopped(_ string) {}

// StackTop implements StackListener — shows the page that is now on top.
func (p *Pages) StackTop(name string) {
	p.SwitchToPage(name)
}
