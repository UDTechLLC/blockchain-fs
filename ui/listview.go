package main

import (
	"strconv"

	"github.com/leedark/ui"
)

// ListView Model
type Filesystem struct {
	Index      int
	Origin     string
	OriginPath string
	Type       int
	Mountpoint string
}

type FilesystemDB struct {
	Filesystems []Filesystem
}

// implement the TableModelHandler interface

func (db *FilesystemDB) NumColumns(m *ui.TableModel) int {
	return 5
}

func (db *FilesystemDB) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (db *FilesystemDB) NumRows(m *ui.TableModel) int {
	return len(db.Filesystems)
}

func (db *FilesystemDB) CellValue(m *ui.TableModel, row int, col int) interface{} {
	value := &db.Filesystems[row]
	switch col {
	case 0:
		return strconv.Itoa(value.Index)
	case 1:
		return value.Origin
	case 2:
		return value.OriginPath
	case 3:
		return strconv.Itoa(value.Type)
	case 4:
		return value.Mountpoint
	}
	return nil
}

func (db *FilesystemDB) SetCellValue(*ui.TableModel, int, int, interface{}) {
	// TODO
}

func (db *FilesystemDB) FindByOrigin(origin string) bool {
	for _, value := range db.Filesystems {
		if value.Origin == origin {
			return true
		}
	}
	return false
}
