package printer

import (
	"fmt"
	"io"
	"reflect"
	"statusmatectl/api"
	"text/tabwriter"
	"time"
)

func PrintAsTable[T any](writer io.Writer, paginated *api.Paginated[T], config *PrintTableConfig) error {
	tw := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)

	if config.PrintBlockTotal {
		_, err := fmt.Fprintf(tw, "total %d\n\n", paginated.Count)
		if err != nil {
			return err
		}
	}

	if len(paginated.Results) == 0 {
		fmt.Fprintln(writer, "No results to display.")
		return nil
	}

	elemType := reflect.TypeOf(paginated.Results[0])

	var fieldIndexes []int
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if tag := field.Tag.Get("tab"); tag != "" {
			fmt.Fprintf(tw, "%s\t", tag)
			fieldIndexes = append(fieldIndexes, i)
		}
	}

	fmt.Fprintln(tw)

	for _, result := range paginated.Results {
		val := reflect.ValueOf(result)
		for _, i := range fieldIndexes {
			fieldVal := val.Field(i)

			if fieldVal.Kind() == reflect.Pointer && !fieldVal.IsNil() {
				fieldVal = fieldVal.Elem()
			}

			switch v := fieldVal.Interface().(type) {
			case time.Time:
				fmt.Fprintf(tw, "%s\t", v.Format("02 Jan 15:04"))
			case nil:
				fmt.Fprint(tw, "\t")
			default:
				fmt.Fprintf(tw, "%v\t", fieldVal)
			}
		}
		fmt.Fprintln(tw)
	}

	tw.Flush()

	return nil
}
