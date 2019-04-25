package main

import (
	"fmt"
	"time"
	"./queue"
	"./active_mq"
)

func main () {
	listenQueue()

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