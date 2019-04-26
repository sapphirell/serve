package model

import (
	"database/sql"
	"fmt"
	"strconv"
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

func (m *CprOrdersModel) GetOrderDetail(orderId int64) error {
	m.construct()

	queryString := "SELECT id,record_id,notify_url,ext FROM " + table + " WHERE id = ?"
	err := m.Db.QueryRow(queryString, orderId).Scan(&m.Id, &m.RecordId, &m.NotifyUrl, &m.Ext)

	return err
}

func (m *CprOrdersModel) SaveOrderResult(renderTime float64, resultVideoUrl string, notifyResult int)  {
	m.construct()
	var queryString string

	queryString  = "UPDATE " + table + " SET `render_time` = '" + strconv.FormatFloat(renderTime,'f',-1,64) + "' "
	queryString += " ,`result_video_url` = '" + resultVideoUrl + "' "
	queryString += " ,`notify_result` = '" + strconv.FormatInt(int64(notifyResult),10) + "' "

	_ , err 	:= m.Db.Exec(queryString)
	if err 		!= nil {
		fmt.Println(err)
	}
}