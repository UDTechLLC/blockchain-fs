package main

import (
	"fmt"
	"strconv"

	"github.com/leedark/ui"
)

// Wallet Model
type Wallet struct {
	Index   int
	Address string
	Credit  int
}

type WalletDB struct {
	Wallets []Wallet
}

// implement the TableModelHandler interface

func (db *WalletDB) NumColumns(m *ui.TableModel) int {
	return 3
}

func (db *WalletDB) ColumnType(m *ui.TableModel, col int) ui.TableModelColumnType {
	return ui.StringColumn
}

func (db *WalletDB) NumRows(m *ui.TableModel) int {
	return len(db.Wallets)
}

func (db *WalletDB) CellValue(m *ui.TableModel, row int, col int) interface{} {
	if row >= len(db.Wallets) {
		fmt.Println("Error: try to get row %d when len is %d", row, len(db.Wallets))
		return nil
	}
	value := &db.Wallets[row]
	switch col {
	case 0:
		return strconv.Itoa(row + 1)
	case 1:
		return value.Address
	case 2:
		return strconv.Itoa(value.Credit)
	}
	return nil
}

func (db *WalletDB) SetCellValue(*ui.TableModel, int, int, interface{}) {
	// TODO
}
