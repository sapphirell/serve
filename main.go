package main

import (
	"./active_mq"
	"./queue"
	"fmt"
	"github.com/Unknwon/goconfig"
	"time"
)

func main () {
	listenQueue()

	return
}

func listenQueue() {
	configPath 				:= "./config.ini"
	config, load_conf_err 	:= goconfig.LoadConfigFile(configPath)
	if load_conf_err != nil {
		fmt.Println(load_conf_err)
	}
	cprQueue, _ := config.GetValue("stomp", "cpr_queue");


	ins := active_mq.ActiveMQInstance{}
	ins.Init()
	ins.Sub(cprQueue)
	var userTask string
	var t queue.CallbackTask

	ticker := time.NewTicker(10 * time.Millisecond)


	for _ = range ticker.C {

		userTask = ins.Get(cprQueue)
		t = queue.CallbackTask{
			Timeout		: 1,
			MaxRepeat 	: 5,
		}
		fmt.Println(userTask)
		//start user task
		t.StartBy(userTask, &ins, cprQueue)

	}
	defer ins.Conn.Disconnect()
}