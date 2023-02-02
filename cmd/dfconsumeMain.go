package main

import (
	"encoding/json"
	"fmt"
	"log"
	"statistics/tools/kafka"
	"statistics/tools/request"
	"time"
)

type ProxyHttp struct {
	Url     string `json:"url"`
	Method  string `json:"method"`
	Headers map[string]string
	Body    string `json:"body"`
}
type DfConsumer struct {
}

func (h *DfConsumer) consume(data []byte) {
	if data == nil {
		return
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("[error]:%s\n", err)
		}
	}()
	var proxyHttp ProxyHttp
	err := json.Unmarshal(data, &proxyHttp)
	if err != nil {
		log.Printf("[error]: json.Unmarshal():%s\n", err.Error())
		return
	}
	if proxyHttp.Url == "" || proxyHttp.Method == "" {
		log.Printf("[error]: dataError:%s\n", string(data))
		return
	}
	beginTime := time.Now().UnixMilli()
	//发起请求
	respBody, err := request.Do(proxyHttp.Method, proxyHttp.Url, proxyHttp.Headers, []byte(proxyHttp.Body))
	//记录请求日志
	logTime := time.Now().Format("2006-01-02 15:04:05")
	jsonFormat := map[string]interface{}{
		"url":          proxyHttp.Url,
		"method":       proxyHttp.Method,
		"request_time": float64(time.Now().UnixMilli()-beginTime) / 1000,
		"log_time":     logTime,
		"resp_body":    string(respBody),
		"err":          fmt.Sprintf("%v", err),
	}
	jsonLog, err := json.Marshal(jsonFormat)
	if err != nil {
		log.Println("write log json.Marshal() failed!", string(jsonLog))
		return
	}
}

func main() {
	//启动消费服务
	//Kafka topic配置
	var (
		brokers = []string{"10.66.224.160:9092"}
		topic   = "iclient_statistics_huoshan"
		groupId = "backend_prod"
	)
	for {
		handler := &DfConsumer{}
		consumer := kafka.NewConsumer("consumer-biz", brokers, topic, groupId)
		consumer.SetResultCallback(handler.consume)
		consumer.SetConcurrency(100) //设置并发数量
		consumer.Run()
		time.Sleep(time.Second * 5)
	}
}
