package stringutil

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

// TableForSlice prints a table for a given slice.
// It will infer column names from the struct fields.
// If it is a mixed array (i.e. []interface{}) it will probably panic.
func TableForSlice[A any](wr io.Writer, collection []A) error {
	cv := reflect.ValueOf(collection)
	for cv.Kind() == reflect.Ptr {
		cv = cv.Elem()
	}
	ct := cv.Type()
	for ct.Kind() == reflect.Ptr || ct.Kind() == reflect.Slice {
		ct = ct.Elem()
	}

	columns := make([]string, ct.NumField())
	for index := 0; index < ct.NumField(); index++ {
		columns[index] = ct.Field(index).Name
	}

	var rows [][]string
	var rowValue reflect.Value
	for row := 0; row < cv.Len(); row++ {
		rowValue = cv.Index(row)
		rowValues := make([]string, ct.NumField())
		for fieldIndex := 0; fieldIndex < ct.NumField(); fieldIndex++ {
			rowValues[fieldIndex] = fmt.Sprintf("%v", rowValue.Field(fieldIndex).Interface())
		}
		rows = append(rows, rowValues)
	}

	return Table(wr, columns, rows)
}

// Table writes a table to a given writer.
func Table(wr io.Writer, columns []string, rows [][]string) error {
	if len(columns) == 0 {
		return nil
	}
	write := func(str string) error {
		_, writeErr := io.WriteString(wr, str)
		return writeErr
	}

	/* begin establish max widths of columns */
	maxWidths := make([]int, len(columns))
	for index, columnName := range columns {
		maxWidths[index] = stringWidth(columnName)
	}

	var width int
	for _, cols := range rows {
		for index, columnValue := range cols {
			width = stringWidth(columnValue)
			if maxWidths[index] < width {
				maxWidths[index] = width
			}
		}
	}
	/* end establish max widths of columns */

	var err error

	/* draw top of column row */
	if err = write(tableTopLeft); err != nil {
		return err
	}
	for index := range columns {
		if err = write(repeat(tableHorizBar, maxWidths[index])); err != nil {
			return err
		}
		if isNotLast(index, columns) {
			if err = write(tableTopSep); err != nil {
				return err
			}
		}
	}
	if err = write(tableTopRight); err != nil {
		return err
	}
	if err = write(newLine); err != nil {
		return err
	}
	/* end draw top of column row */

	/* draw column names */
	if err = write(tableVertBar); err != nil {
		return err
	}
	for index, columnLabel := range columns {
		if err = write(padRight(columnLabel, maxWidths[index])); err != nil {
			return err
		}
		if isNotLast(index, columns) {
			if err = write(tableVertBar); err != nil {
				return err
			}
		}
	}
	if err = write(tableVertBar); err != nil {
		return err
	}
	if err = write(newLine); err != nil {
		return err
	}
	/* end draw column names */

	/* draw bottom of column row */
	if err = write(tableMidLeft); err != nil {
		return err
	}
	for index := range columns {
		if err = write(repeat(tableHorizBar, maxWidths[index])); err != nil {
			return err
		}
		if isNotLast(index, columns) {
			if err = write(tableMidSep); err != nil {
				return err
			}
		}
	}
	if err = write(tableMidRight); err != nil {
		return err
	}
	if err = write(newLine); err != nil {
		return err
	}
	/* end draw bottom of column row */

	/* draw rows */
	for _, row := range rows {
		if err = write(tableVertBar); err != nil {
			return err
		}
		for index, column := range row {
			if err = write(padRight(column, maxWidths[index])); err != nil {
				return err
			}
			if isNotLast(index, columns) {
				if err = write(tableVertBar); err != nil {
					return err
				}
			}
		}
		if err = write(tableVertBar); err != nil {
			return err
		}
		if err = write(newLine); err != nil {
			return err
		}
	}
	/* end draw rows */

	/* draw footer */
	if err = write(tableBottomLeft); err != nil {
		return err
	}
	for index := range columns {
		if err = write(repeat(tableHorizBar, maxWidths[index])); err != nil {
			return err
		}
		if isNotLast(index, columns) {
			if err = write(tableBottomSep); err != nil {
				return err
			}
		}
	}
	if err = write(tableBottomRight); err != nil {
		return err
	}
	if err = write(newLine); err != nil {
		return err
	}
	/* end draw footer */
	return nil
}

const (
	tableTopLeft     = "┌"
	tableTopRight    = "┐"
	tableBottomLeft  = "└"
	tableBottomRight = "┘"
	tableMidLeft     = "├"
	tableMidRight    = "┤"
	tableVertBar     = "│"
	tableHorizBar    = "─"
	tableTopSep      = "┬"
	tableBottomSep   = "┴"
	tableMidSep      = "┼"
	newLine          = "\n"
)

func stringWidth(value string) (width int) {
	var runeWidth int
	for _, c := range value {
		runeWidth = utf8.RuneLen(c)
		if runeWidth > 1 {
			width += 2
		} else {
			width++
		}
	}
	return
}

func repeat(str string, count int) string {
	return strings.Repeat(str, count)
}

func padRight(value string, width int) string {
	valueWidth := stringWidth(value)
	spaces := width - valueWidth
	if spaces == 0 {
		return value
	}
	return value + strings.Repeat(" ", spaces)
}

func isNotLast(index int, values []string) bool {
	return index < (len(values) - 1)
}
