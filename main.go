package main

import (
	"fmt"
	"./model"
	"./active_mq"
	"./queue"
	"time"
)

func main () {
	//listenQueue()
	//d := model.DbLinker{}
	//d.Init()
	//d.DB.Exec("insert into `vep_request_test` set data = '123' ")
	m := model.CprOrdersModel{}
	m.GetOrderDetail(1)

	fmt.Println(m.Ext)
	return
}

func listenQueue() {
	ins := active_mq.ActiveMQInstance{}
	ins.Init()
	ins.Sub("/cpr_queue")
	var msg string
	var t queue.CallbackTask

	ticker := time.NewTicker(500 * time.Millisecond)

	for _ = range ticker.C {
		//fmt.Println(time.Now())
		msg = ins.Get("/cpr_queue")
		t = queue.CallbackTask{
			Timeout		: 5,
			MaxRepeat 	: 5,
		}
		fmt.Println(msg)
		t.StartBy(msg)



	}
	defer ins.Conn.Disconnect()
}