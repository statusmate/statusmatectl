package printer

import "fmt"

type PrintTableFormatType = string

const (
	PrintTableFormatJSON  PrintTableFormatType = "json"
	PrintTableFormatTable PrintTableFormatType = "table"
	PrintTableFormatList  PrintTableFormatType = "list"
)

func ValidatePrintTableFormat(format string) error {
	switch format {
	case PrintTableFormatJSON, PrintTableFormatTable, PrintTableFormatList:
		return nil
	default:
		return fmt.Errorf("unsupported format: %s, choose 'json' or 'table'", format)
	}
}

type PrintTableConfig struct {
	PrintBlockTotal bool
	Format          PrintTableFormatType
}

func NewPrintTableConfig() *PrintTableConfig {
	return &PrintTableConfig{
		PrintBlockTotal: true,
		Format:          PrintTableFormatTable,
	}
}
