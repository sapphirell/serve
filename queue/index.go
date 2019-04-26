package queue

import (
    "../active_mq"
    "bytes"
    json2 "encoding/json"
    "fmt"
    "github.com/tidwall/gjson"
    "net/http"
    "time"
    "../model"
)

type CallbackTask struct {
    OriData     string //Json Data
    UserData    *UserData
    Timeout     int64
    MaxRepeat   int32 // Discard task if over this number
}
//save to queue
type UserData struct {
    OrderId        int64    `json:"orderId"`
    Repeated       int64    `json:"repeated"`
    NotifyUrl      string   `json:"notifyUrl"`
    StatusCode     int64    `json:"statusCode"`// 0 == ok
    RenderTime     float64  `json:"renderTime"`
    RecordId       string   `json:"recordId"`//User's record id
    ResultVideoUrl string   `json:"resultVideoUrl"`//Render video save path
}

//post to user
type PostJson struct {
    RecordId       string  `json:"record_id"`
    ResultVideoUrl string  `json:"result_video_url"`
    RenderTime     float64 `json:"render_time"`
    StatusCode     int64   `json:"status_code"`
}

//channel := make(chan os.Signal)
//signal.Notify(channel, os.Interrupt, os.Kill)
//<-channel
func (t *CallbackTask) StartBy(oriData string, mq *active_mq.ActiveMQInstance, queueName string) {
    t.OriData   = oriData
    flag        := t.init()
    if flag {
        t.TaskPoster(mq, queueName)
    }else {
        println("丢弃任务")
    }

}

//construct `UserData` from queue or db
func (t *CallbackTask) init() bool {
    if t.OriData == "" {
        return false
    }
    c           := model.CprOrdersModel{}
    tmpUserData := UserData{}
    tmpUserData.Repeated          = gjson.Get(t.OriData, "repeated").Int() + 1
    tmpUserData.OrderId           = gjson.Get(t.OriData, "orderId").Int()
    tmpUserData.StatusCode        = gjson.Get(t.OriData, "statusCode").Int()
    tmpUserData.RenderTime        = gjson.Get(t.OriData, "renderTime").Float()
    tmpUserData.ResultVideoUrl    = gjson.Get(t.OriData, "resultVideoUrl").String()
    if int(t.MaxRepeat) < int(tmpUserData.Repeated) {
        //save data
        c.SaveOrderResult(tmpUserData.RenderTime, tmpUserData.ResultVideoUrl, 2)
        return false
    }
    if tmpUserData.Repeated == 1 {

        err := c.GetOrderDetail(tmpUserData.OrderId)
        if err != nil {
            fmt.Println(err)
            return false
        }

        tmpUserData.NotifyUrl = c.NotifyUrl
        tmpUserData.RecordId  = c.RecordId

    } else {
        tmpUserData.NotifyUrl         = gjson.Get(t.OriData, "notifyUrl").String()
        tmpUserData.RecordId          = gjson.Get(t.OriData, "recordId").String()
    }

    t.UserData = &tmpUserData
    return true
}

func (t *CallbackTask) TaskPoster(mq *active_mq.ActiveMQInstance,queueName string) {
    timeout         := make(chan bool, 1)
    requestResult   := make(chan bool, 0)
    requestFailed   := make(chan bool, 0)


    go func() {
        time.Sleep(time.Duration(t.Timeout) * time.Second)
        timeout <- true
    }()

    go func() {
        postData := PostJson{
            RecordId:       t.UserData.RecordId,
            ResultVideoUrl: t.UserData.ResultVideoUrl,
            RenderTime:     t.UserData.RenderTime,
            StatusCode:     t.UserData.StatusCode,
        }
        jsonByte, _ := json2.Marshal(postData)
        req, err    := http.NewRequest("POST", t.UserData.NotifyUrl, bytes.NewBuffer(jsonByte))
        req.Header.Set("Content-Type", "application/json")

        client      := &http.Client{}
        resp, err   := client.Do(req)

        if err != nil {
            requestFailed   <- true

        } else {

            if resp != nil {
                c := model.CprOrdersModel{}
                c.SaveOrderResult(t.UserData.RenderTime, t.UserData.ResultVideoUrl, 1)
                requestResult <- true
            }

            _ = resp.Body.Close()
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
            t.repeatQueue(mq, queueName)
        }
    case <-requestFailed:
        {
            fmt.Println("Request failed!") //TODO:重新加入队列
            t.repeatQueue(mq, queueName)
        }

    }
}

func (t *CallbackTask) repeatQueue(mq *active_mq.ActiveMQInstance,queueName string) {
    userData, _ := json2.Marshal(t.UserData)
    fmt.Println(string(queueName))
    mq.Push(queueName, string(userData))
}

