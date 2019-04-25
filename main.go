package main

import (
	"fmt"
	"time"
	"./src"
)

func main () {
	listenQueue()

	return
}

func listenQueue() {
	ins := src.ActiveMQInstance{}
	ins.Init()
	ins.Sub("/cpr_queue")
	var msg string
	var t src.CallbackTask

	ticker := time.NewTicker(500 * time.Millisecond)

	for _ = range ticker.C {
		//fmt.Println(time.Now())
		msg = ins.Get("/cpr_queue")
		t = src.CallbackTask{
			Timeout		: 5,
			MaxRepeat 	: 5,
		}
		fmt.Println(msg)
		t.StartBy(msg)


	}
	defer ins.Conn.Disconnect()
}