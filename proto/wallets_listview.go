package main

import (
	"strconv"

	"github.com/leedark/ui"
)

// Wallet Model
type Wallet struct {
	Index   int
	Address string
}

type WalletDB struct {
	Wallets []Wallet
}

// implement the TableModelHandler interface

func (db *WalletDB) NumColumns(m *ui.TableModel) int {
	return 2
}

func (db *WalletDB) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (db *WalletDB) NumRows(m *ui.TableModel) int {
	return len(db.Wallets)
}

func (db *WalletDB) CellValue(m *ui.TableModel, row int, col int) interface{} {
	value := &db.Wallets[row]
	switch col {
	case 0:
		return strconv.Itoa(row + 1)
	case 1:
		return value.Address
	}
	return nil
}

func (db *WalletDB) SetCellValue(*ui.TableModel, int, int, interface{}) {
	// TODO
}
