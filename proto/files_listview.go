package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/leedark/ui"
)

// File Model
type File struct {
	Index     int
	RaftIndex int
	Name      string
	Timestamp time.Time
}

type FileDB struct {
	Files []File
}

// implement the TableModelHandler interface

func (db *FileDB) NumColumns(m *ui.TableModel) int {
	return 4
}

func (db *FileDB) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (db *FileDB) NumRows(m *ui.TableModel) int {
	return len(db.Files)
}

func (db *FileDB) CellValue(m *ui.TableModel, row int, col int) interface{} {
	if row >= len(db.Files) {
		fmt.Println("Error: try to get row %d when len is %d", row, len(db.Files))
		return nil
	}
	value := &db.Files[row]
	switch col {
	case 0:
		return strconv.Itoa(row + 1)
	case 1:
		return strconv.Itoa(value.RaftIndex)
	case 2:
		return value.Name
	case 3:
		return value.Timestamp.Format(time.RFC1123)
	}
	return nil
}

func (db *FileDB) SetCellValue(*ui.TableModel, int, int, interface{}) {
	// TODO
}
