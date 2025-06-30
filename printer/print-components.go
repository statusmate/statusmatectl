package printer

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/list"
	"io"
	"statusmatectl/api"
)

// PrintComponents prints a table of components and their nested components.
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

func PrintComponentsAsList(w io.Writer, paginated *api.Paginated[api.Component]) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	components := paginated.Results

	componentMap := make(map[int]*api.Component)
	for i := range components {
		if components[i].ID != nil {
			componentMap[*components[i].ID] = &components[i]
		}
	}

	for _, comp := range components {
		if comp.Parent == nil {
			printComponentTree(l, &comp, componentMap, 0)
		}
	}

	if _, err := w.Write([]byte(l.Render())); err != nil {
		fmt.Println("Error writing to writer:", err)
	}

	_, err := fmt.Fprint(w, "\n")
	if err != nil {
		fmt.Println("Error writing to writer:", err)
	}

	return nil
}

func printComponentTree(l list.Writer, comp *api.Component, componentMap map[int]*api.Component, level int) {
	name := comp.Name

	l.AppendItem(name)

	l.Indent()

	for _, child := range componentMap {
		if child.Parent != nil && *child.Parent == *comp.ID {
			printComponentTree(l, child, componentMap, level+1)
		}
	}

	l.UnIndent()
}
