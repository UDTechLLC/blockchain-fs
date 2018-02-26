package main

import (
	"math/rand"
	"strconv"

	"github.com/icrowley/fake"
	"github.com/leedark/ui"
)

// ListView Model
type Filesystem struct {
	Index int
	Name  string
	Path  string
}

type FilesystemDB struct {
	Filesystems []Filesystem
}

// implement the TableModelHandler interface

func (db *FilesystemDB) NumColumns(m *ui.TableModel) int {
	return 3
}

func (db *FilesystemDB) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (db *FilesystemDB) NumRows(m *ui.TableModel) int {
	return len(db.Filesystems)
}

func (db *FilesystemDB) CellValue(m *ui.TableModel, row int, col int) interface{} {
	prod := &db.Filesystems[row]
	switch col {
	case 0:
		return strconv.Itoa(prod.Index)
	case 1:
		return prod.Name
	case 2:
		return prod.Path
	}
	return nil
}

func (db *FilesystemDB) SetCellValue(*ui.TableModel, int, int, interface{}) {
	// TODO
}

func RandomFilesystem() Filesystem {
	fs := Filesystem{}
	fs.Index = 0 + rand.Intn(10)
	fs.Name = fake.FirstName()
	fs.Path = fake.LastName()
	return fs
}
