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
    NotifyTime     int64    `json:"notifyTime"`
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
    var flag int
    //var s string
    t.OriData   = oriData
    flag = t.init()
    _ =  t.parseFlag(flag, mq, queueName)
}

//construct `UserData` from queue or db
func (t *CallbackTask) init() int {
    if t.OriData == "" {
        return 1
    }
    c           := model.CprOrdersModel{}
    tmpUserData := UserData{}
    tmpUserData.Repeated          = gjson.Get(t.OriData, "repeated").Int()
    tmpUserData.OrderId           = gjson.Get(t.OriData, "orderId").Int()
    tmpUserData.StatusCode        = gjson.Get(t.OriData, "statusCode").Int()
    tmpUserData.RenderTime        = gjson.Get(t.OriData, "renderTime").Float()
    tmpUserData.ResultVideoUrl    = gjson.Get(t.OriData, "resultVideoUrl").String()
    tmpUserData.NotifyTime        = gjson.Get(t.OriData, "notifyTime").Int()

    if int(t.MaxRepeat) <= int(tmpUserData.Repeated) {
        //save data
        c.SaveOrderResult(tmpUserData.RenderTime, tmpUserData.ResultVideoUrl, 2)
        return 2
    }
    if tmpUserData.Repeated == 0 {

        err := c.GetOrderDetail(tmpUserData.OrderId)
        if err != nil {
            fmt.Println(t.OriData)
            return 4
        }

        tmpUserData.NotifyUrl  = c.NotifyUrl
        tmpUserData.RecordId   = c.RecordId
        tmpUserData.NotifyTime = time.Now().Unix()

    } else {
        tmpUserData.NotifyUrl   = gjson.Get(t.OriData, "notifyUrl").String()
        tmpUserData.RecordId    = gjson.Get(t.OriData, "recordId").String()
    }

    t.UserData = &tmpUserData

    if tmpUserData.NotifyTime > time.Now().Unix() {
        return 3
    }
    return 0
}

func (t *CallbackTask) TaskPoster(mq *active_mq.ActiveMQInstance,queueName string) {
    timeout         := make(chan bool, 1)
    requestResult   := make(chan bool, 0)
    requestFailed   := make(chan bool, 0)
    tmpUserData     := t.UserData
    tmpUserData.Repeated += 1
    t.UserData      = tmpUserData

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
            t.repeatQueue(mq, queueName, true)
        }
    case <-requestFailed:
        {
            fmt.Println("Request failed!") //TODO:重新加入队列
            t.repeatQueue(mq, queueName, true)
        }

    }
}

func (t *CallbackTask) repeatQueue(mq *active_mq.ActiveMQInstance, queueName string, interval bool) {
    tmp := t.UserData
    if interval {
        tmp.NotifyTime += 5 //after 5 sec
    }
    userData, _ := json2.Marshal(tmp)
    //fmt.Println(string(queueName))
    mq.Push(queueName, string(userData))
}

func (t *CallbackTask) parseFlag(flag int, mq *active_mq.ActiveMQInstance,queueName string) string {
    if flag == 0 {
        t.TaskPoster(mq, queueName)
        return "ok"
    }
    if flag == 1 {
        return "Empty JSON"
    }
    if flag == 2 {
        return "over Max Repeat"
    }
    if flag == 3 {
        t.repeatQueue(mq, queueName, false)
        return "Time not come"
    }
    return ""
}