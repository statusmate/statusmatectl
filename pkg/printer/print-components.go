package printer

import (
	"fmt"
	"io"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/statusmate/statusmatectl/pkg/api"
)

func PrintComponents(w io.Writer, paginated *api.Paginated[api.Component], config *PrintTableConfig) error {
	if config.Format == PrintTableFormatJSON {
		return PrintAsJSON(w, paginated, config)
	}

	if config.Format == PrintTableFormatList {
		return PrintComponentsAsList(w, paginated)
	}

	if config.Format == PrintTableFormatTable {
		return PrintAsTable(w, paginated, config)
	}

	return nil
}

func buildComponentMap(components []api.Component) map[int]*api.Component {
	m := make(map[int]*api.Component, len(components))
	for i := range components {
		if components[i].ID != nil {
			m[*components[i].ID] = &components[i]
		}
	}
	return m
}

func renderList(w io.Writer, l list.Writer) error {
	if _, err := fmt.Fprintln(w, l.Render()); err != nil {
		return err
	}
	return nil
}

func PrintComponentsAsList(w io.Writer, paginated *api.Paginated[api.Component]) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	componentMap := buildComponentMap(paginated.Results)

	for _, comp := range paginated.Results {
		if comp.Parent == nil {
			printComponentTree(l, &comp, componentMap, 0)
		}
	}

	return renderList(w, l)
}

func PrintComponentStatusTree(w io.Writer, paginated *api.Paginated[api.Component], reasons map[int][]string) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	componentMap := buildComponentMap(paginated.Results)

	for _, comp := range paginated.Results {
		if comp.Parent == nil {
			printStatusTree(l, &comp, componentMap, reasons)
		}
	}

	return renderList(w, l)
}

func impactIcon(impact api.ImpactType) string {
	switch impact {
	case api.ImpactTypeOperational:
		return "🟢"
	case api.ImpactTypeUnderMaintenance:
		return "🔧"
	case api.ImpactTypeDegradedPerformance:
		return "🟡"
	case api.ImpactTypePartialOutage:
		return "🟠"
	case api.ImpactTypeMajorOutage:
		return "🔴"
	default:
		return "⚪"
	}
}

func printComponentTree(l list.Writer, comp *api.Component, componentMap map[int]*api.Component, level int) {
	l.AppendItem(fmt.Sprintf("%s %s", impactIcon(comp.Impact), comp.Name))

	l.Indent()
	for _, child := range componentMap {
		if child.Parent != nil && *child.Parent == *comp.ID {
			printComponentTree(l, child, componentMap, level+1)
		}
	}
	l.UnIndent()
}

func printStatusTree(l list.Writer, comp *api.Component, componentMap map[int]*api.Component, reasons map[int][]string) {
	l.AppendItem(fmt.Sprintf("%s %s", impactIcon(comp.Impact), comp.Name))

	l.Indent()

	if comp.ID != nil {
		for _, reason := range reasons[*comp.ID] {
			l.AppendItem(reason)
		}
		for _, child := range componentMap {
			if child.Parent != nil && *child.Parent == *comp.ID {
				printStatusTree(l, child, componentMap, reasons)
			}
		}
	}

	l.UnIndent()
}
