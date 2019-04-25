package model

import (
	"database/sql"
	"fmt"
)

const table = "vep_cpr_orders"

type CprOrdersModel struct {
	linker    DbLinker
	Db        *sql.DB
	Id        int
	RecordId  string
	NotifyUrl string
	Ext       string
}

func (m *CprOrdersModel) construct() {
	d := DbLinker{}
	d.Init()

	m.linker, m.Db = d, d.DB
}

func (m *CprOrdersModel) Exec(query string, args ...interface{}) (sql.Result, error) {
	var res sql.Result
	var err error
	if args == nil {
		res, err = m.linker.DB.Exec(query)
	} else {
		res, err = m.linker.DB.Exec(query, args)
	}

	return res, err
}

func (m *CprOrdersModel) GetOrderDetail(orderId int) {
	m.construct()

	queryString := "SELECT id,record_id,notify_url,ext FROM " + table + " WHERE id = ?"
	err := m.Db.QueryRow(queryString, orderId).Scan(&m.Id, &m.RecordId, &m.NotifyUrl, &m.Ext)

	if err != nil {
		fmt.Println(err)
	}
}
