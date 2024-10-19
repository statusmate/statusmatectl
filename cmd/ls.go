package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"text/tabwriter"
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Ls command",
	RunE:  lsCmdF,
}

func init() {
	RootCmd.AddCommand(LsCmd)
}

func lsCmdF(command *cobra.Command, args []string) error {

	items := []Item{
		{"Apple", "$1", 10},
		{"Banana", "$0.5", 20},
		{"Cherry", "$2", 15},
	}

	// Вызываем функцию для печати в стандартный вывод
	PrintTable(os.Stdout, items)
	return nil
}

type Item struct {
	Name  string `json:"name"  tab:"10"`
	Price string `json:"price" tab:"10"`
	Count int    `json:"count" tab:"5"`
}

// PrintTable выводит массив элементов в виде таблицы в указанный Writer
func PrintTable(w io.Writer, items []Item) {
	// Создаем новый tabwriter
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Печатаем заголовок
	fmt.Fprintln(tw, "Name\tPrice\tCount\t")

	// Печатаем каждую строку с данными
	for _, item := range items {
		fmt.Fprintf(tw, "%s\t%s\t%d\t\n", item.Name, item.Price, item.Count)
	}

	// Завершаем вывод
	tw.Flush()
}
