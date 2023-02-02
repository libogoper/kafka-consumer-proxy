package main

import (
	"encoding/json"
	"fmt"
	"log"
	_ "net/http/pprof"
	"statistics/tools/kafka"
	"statistics/tools/request"
	"time"
)

const consumeApiGateway = "http://local.test.com/consume"

type StatConsumer struct {
}

func (*StatConsumer) consume(data []byte) {
	if data == nil {
		return
	}
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("[error]:%s\n", err)
		}
	}()
	beginTime := time.Now().UnixMilli()
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	//发起请求
	respBody, err := request.Do("POST", consumeApiGateway, headers, data)
	//记录请求日志
	logTime := time.Now().Format("2006-01-02 15:04:05")
	jsonFormat := map[string]interface{}{
		"url":          consumeApiGateway,
		"method":       "POST",
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
		topic   = "iclient_statistics"
		groupId = "connect-sink-statistic-php-report-connector"
	)
	for {
		handler := &StatConsumer{}
		consumer := kafka.NewConsumer("consumer-biz-stat", brokers, topic, groupId)
		consumer.SetResultCallback(handler.consume)
		consumer.SetConcurrency(300) //设置并发数量
		consumer.Run()
		time.Sleep(time.Second * 5)
	}
}
