package src

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/go-stomp/stomp"
)

//Get task by queue name
type ActiveMQInstance struct {
	Conn         *stomp.Conn
	Subscription map[string]*stomp.Subscription
	Test         map[string]string
}

func (ins *ActiveMQInstance) Init() {
	configPath 				:= "./config.ini"
	config, load_conf_err 	:= goconfig.LoadConfigFile(configPath)
	if load_conf_err != nil {
		fmt.Println(load_conf_err)
	}
	//禁用timeOut
	var options []func(*stomp.Conn) error = []func(*stomp.Conn) error{
		stomp.ConnOpt.HeartBeat(0, 0),
	}

	activeMqHost, _ := config.GetValue("stomp", "host");
	conn, err 		:= stomp.Dial("tcp", activeMqHost, options...)
	if err != nil {
		fmt.Println(err)
	}
	ins.Conn = conn
}

func (ins *ActiveMQInstance) Sub(queueName string) {
	sub, err := ins.Conn.Subscribe(queueName, stomp.AckAuto)
	if err != nil {
		fmt.Println(err)
	}

	tmp 			 := make(map[string]*stomp.Subscription)
	tmp[queueName] 	 = sub
	ins.Subscription = tmp
}

func (ins *ActiveMQInstance) Get(queueName string) string {
	msg := <-ins.Subscription[queueName].C

	return string(msg.Body)
}

func (ins * ActiveMQInstance) Push(queueName string, body string) {
	err := ins.Conn.Send( queueName, "text/plain", []byte(body))
	if err != nil {
		println(err)
	}

}