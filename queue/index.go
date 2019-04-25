package queue

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
	"time"
)

type CallbackTask struct {
	UserData       string //Json Data

	Timeout        int64
	MaxRepeat      int32 // Discard task if over this number

	Repeated       int32
	NotifyUrl      string
	StatusCode     int64 // 0 == ok
	RenderTime     float64
	RecordId       string //User's record id
	ResultVideoUrl string //Render video save path
}

type PostJson struct {
	RecordId       string  `json:"record_id"`
	ResultVideoUrl string  `json:"result_video_url"`
	RenderTime     float64 `json:"render_time"`
	StatusCode     int64   `json:"status_code"`
}

//channel := make(chan os.Signal)
//signal.Notify(channel, os.Interrupt, os.Kill)
//<-channel
func (t *CallbackTask) StartBy(userData string) int {
	t.UserData  = userData
	t.init()
	t.TaskPoster()
	return 1;
}
func (t *CallbackTask) init() bool {
	if t.UserData == "" {
		return false
	}

	t.Repeated 		 = 5
	t.NotifyUrl		 = gjson.Get(t.UserData, "notifyUrl").String()
	t.StatusCode 	 = gjson.Get(t.UserData, "statusCode").Int()
	t.RenderTime 	 = gjson.Get(t.UserData, "renderTime").Float()
	t.RecordId 		 = gjson.Get(t.UserData, "recordId").String()
	t.ResultVideoUrl = gjson.Get(t.UserData, "resultVideoUrl").String()

	return true
}

func (t *CallbackTask) TaskPoster() {
	timeout 		:= make(chan bool, 1)
	requestResult	:= make(chan bool, 0)
	requestFailed	:= make(chan bool, 0)
	todoAction		:= make(chan string,0)
	go func() {
		time.Sleep(time.Duration(t.Timeout) * time.Second)
		timeout <- true
	}()
	go func() {
		postData := PostJson{
			RecordId:       t.RecordId,
			ResultVideoUrl: t.ResultVideoUrl,
			RenderTime:     t.RenderTime,
			StatusCode:     t.StatusCode,
		}
		jsonByte, _ := json2.Marshal(postData)

		req, err 	:= http.NewRequest("POST", t.NotifyUrl, bytes.NewBuffer(jsonByte))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			todoAction <- t.UserData
			requestFailed <- true

		} else {

			if resp != nil {
				println(resp.StatusCode)
				requestResult <- true
			}

			resp.Body.Close()
		}


	}()

	select {
	case <-requestResult:
		{
			fmt.Println("callBack complete!")
		}
	case <-timeout:
		{
			fmt.Println("callback timeout!")
			//TODO:重新加入队列
		}
	case <-requestFailed:
		{
			fmt.Println( "Request failed!") //TODO:重新加入队列
		}

	}
}


