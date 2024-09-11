package table

import (
	"errors"
	"fmt"
	"github.com/MathiasMantai/gotools/cli"
	"github.com/MathiasMantai/gotools/datastructures"
)

type Table struct {
	ColumnWidth  []int
	Data         [][]string
	DividerWidth int
	WhiteSpace   int
}

func (t *Table) validateTable() error {

	if len(t.Data) == 0 {
		return errors.New("table is empty")
	}

	startLen := len(t.Data[0])

	for i := startLen; i < len(t.Data); i++ {
		if len(t.Data[i]) != startLen {
			return errors.New("all rows must have the same number of columns")
		}
	}

	return nil
}

func (t *Table) getMaxStringLengthPerColumn() {
	tableColumnLength := make([]int, len(t.Data[0]))
	for _, row := range t.Data {
		for j, column := range row {
			if len(column) > tableColumnLength[j] {
				tableColumnLength[j] = len(column)
			}
		}
	}

	t.ColumnWidth = tableColumnLength
}

func (t *Table) getTopAndBottomWidth() {
	t.DividerWidth = datastructures.GetIntSliceSum(t.ColumnWidth) + (len(t.ColumnWidth) * (t.WhiteSpace * 2)) + (len(t.ColumnWidth) + 1)
}

/* PRINT TABLE */
func (t *Table) printTable() {
	fmt.Print(t.getDivider())
	for _, row := range t.Data {
		fmt.Print(t.getRow(row))
		fmt.Print(t.getDivider())
	}
}

func (t *Table) getDivider() string {
	rs := ""

	var i int
	for i < t.DividerWidth {
		rs += "-"
		i++
	}
	rs += "\n"
	return rs
}

func getWhiteSpace(whiteSpace int) string {
	rs := ""
	var i int
	for i < whiteSpace {
		rs += " "
		i++
	}

	return rs
}

func (t *Table) getRow(row []string) string {
	rs := "|"

	for key, value := range row {
		maxWidthForColumn := t.ColumnWidth[key]
		widthCurrentValue := len(value)
		maxWidthDifference := maxWidthForColumn - widthCurrentValue

		rs += getWhiteSpace(t.WhiteSpace+maxWidthDifference) + value + getWhiteSpace(t.WhiteSpace) + "|"
	}
	rs += "\n"
	return rs
}

/*PRINT HEADER */
func (t *Table) printHeader(color string) {
	fmt.Print(t.getDivider())
	fmt.Print(t.getHeaderRow(color))
}

func (t *Table) getHeaderRow(color string) string {
	rs := "|"
	row := t.Data[0]
	for key, value := range row {
		maxWidthForColumn := t.ColumnWidth[key]
		widthCurrentValue := len(value)
		maxWidthDifference := maxWidthForColumn - widthCurrentValue

		rs += getWhiteSpace(t.WhiteSpace+maxWidthDifference) + cli.GetBoldAndColor(value, color, false) + getWhiteSpace(t.WhiteSpace) + "|"
	}
	rs += "\n"
	return rs
}

func (t *Table) Print(tableData [][]string, whiteSpace int) error {
	t.Data = tableData
	t.WhiteSpace = whiteSpace

	if tableError := t.validateTable(); tableError != nil {
		return tableError
	}

	t.getMaxStringLengthPerColumn()
	t.getTopAndBottomWidth()
	t.printTable()
	return nil
}

func (t *Table) PrintWithHeader(tableData [][]string, headerColor string, whiteSpace int) error {
	t.Data = append([][]string{tableData[0]}, tableData[1:]...)
	t.WhiteSpace = whiteSpace

	if tableError := t.validateTable(); tableError != nil {
		return tableError
	}

	t.getMaxStringLengthPerColumn()
	t.getTopAndBottomWidth()
	t.printHeader(headerColor)
	t.Data = datastructures.InterfaceToTwoDStringSlice(datastructures.RemoveSliceValueTwoD(0, datastructures.TwoDStringToInterfaceSlice(t.Data)))

	t.printTable()
	return nil
}
