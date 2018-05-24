package ui

// #include <stdlib.h>
// #include "ui.h"
//
// extern int doTableModelHandlerNumColumns(uiTableModelHandler *, uiTableModel *);
// extern uiTableModelColumnType doTableModelHandlerColumnType(uiTableModelHandler *, uiTableModel *, int);
// extern int doTableModelHandlerNumRows(uiTableModelHandler *, uiTableModel *);
// extern void* doTableModelHandlerCellValue(uiTableModelHandler *, uiTableModel *, int, int);
// extern void doTableModelHandlerSetCellValue(uiTableModelHandler *, uiTableModel *, int, int, void*);
//
// static inline uiTableModelHandler *allocTableModelHandler(void)
// {
// 	uiTableModelHandler *h;
//
// 	h = (uiTableModelHandler *) malloc(sizeof (uiTableModelHandler));
// 	if (h == NULL)		// TODO
// 		return NULL;
//  h->NumColumns = doTableModelHandlerNumColumns;
//  h->ColumnType = doTableModelHandlerColumnType;
//  h->NumRows = doTableModelHandlerNumRows;
//  h->CellValue = doTableModelHandlerCellValue;
//  h->SetCellValue = doTableModelHandlerSetCellValue;
// 	return h;
// }
// static inline void freeTableModelHandler(uiTableModelHandler *h)
// {
// 	free(h);
// }
import "C"

import "unsafe"

// -------------
// TableModelHandler stuff

// no need to lock this; only the GUI thread can access it
var tablemodelhandlers = make(map[*C.uiTableModelHandler]TableModelHandler)

type TableModelHandler interface {
	NumColumns(*TableModel) int
	ColumnType(*TableModel, int) TableModelColumnType
	NumRows(*TableModel) int
	CellValue(*TableModel, int, int) interface{}
	SetCellValue(*TableModel, int, int, interface{})
}

// Note: these must be numerically identical to their libui equivalents.
type TableModelColumnType uint

const (
	StringColumn TableModelColumnType = iota
	ImageColumn
	IntColumn
	ColorColumn
)

func registerTableModeHandler(h TableModelHandler) *C.uiTableModelHandler {
	uh := C.allocTableModelHandler()
	tablemodelhandlers[uh] = h
	return uh
}

func unregisterTableModelHandler(uh *C.uiTableModelHandler) {
	delete(tablemodelhandlers, uh)
	C.freeTableModelHandler(uh)
}

//export doTableModelHandlerNumColumns
func doTableModelHandlerNumColumns(uh *C.uiTableModelHandler, um *C.uiTableModel) C.int {
	h := tablemodelhandlers[uh]
	m := tablemodels[um]
	return C.int(h.NumColumns(m))
}

//export doTableModelHandlerColumnType
func doTableModelHandlerColumnType(uh *C.uiTableModelHandler, um *C.uiTableModel, col C.int) C.uiTableModelColumnType {
	h := tablemodelhandlers[uh]
	m := tablemodels[um]
	typ := h.ColumnType(m, int(col))
	return C.uiTableModelColumnType(typ)
}

//export doTableModelHandlerNumRows
func doTableModelHandlerNumRows(uh *C.uiTableModelHandler, um *C.uiTableModel) C.int {
	h := tablemodelhandlers[uh]
	m := tablemodels[um]
	return C.int(h.NumRows(m))
}

//export doTableModelHandlerCellValue
func doTableModelHandlerCellValue(uh *C.uiTableModelHandler, um *C.uiTableModel, row C.int, col C.int) unsafe.Pointer {
	h := tablemodelhandlers[uh]
	m := tablemodels[um]

	v := h.CellValue(m, int(row), int(col))

	// TODO: handle other column types!
	s := v.(string)
	tmp := C.CString(s)
	out := C.uiTableModelStrdup(tmp)
	freestr(tmp)
	return out
}

//export doTableModelHandlerSetCellValue
func doTableModelHandlerSetCellValue(uh *C.uiTableModelHandler, um *C.uiTableModel, row C.int, col C.int, value unsafe.Pointer) {
	// TODO!!!
}
