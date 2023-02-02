package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

type Consumer struct {
	srvName        string   //服务名称
	topic          string   //topic name
	brokers        []string //broker list
	groupId        string   //consumer group
	queue          chan []byte
	cancelCtx      context.Context
	cancelFunc     context.CancelFunc
	concurrency    int               //并发
	resultCallback func(data []byte) //接收回调数据
}

func (h Consumer) Setup(sess sarama.ConsumerGroupSession) error {
	h.logln(fmt.Sprintf("Setup:%v", sess.Claims()))
	return nil
}

func (h Consumer) Cleanup(sess sarama.ConsumerGroupSession) error {
	h.logln(fmt.Sprintf("Cleanup:%v", sess.Claims()))
	return nil
}

// 接收消费消息
func (h Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 获取消息
	for msg := range claim.Messages() {
		if msg.Partition == 0 && rand.Intn(100) < 2 {
			tm := msg.Timestamp.Format("2006-01-02 15:04:05")
			h.logln(fmt.Sprintf("Message topic:%q partition:%d offset:%d tm:%s", msg.Topic, msg.Partition, msg.Offset, tm))
		}
		h.queue <- msg.Value
		// 将消息标记为已使用
		sess.MarkMessage(msg, "")
	}
	return nil
}
func (h *Consumer) SetConcurrency(n int) {
	h.concurrency = n
}
func (h *Consumer) SetResultCallback(callback func(data []byte)) {
	h.resultCallback = callback
}
func (h Consumer) logln(msg string) {
	log.Printf("%s|%s\n", h.srvName, msg)
}
func (h Consumer) logFatalln(msg string) {
	log.Fatalf("%s|%s\n", h.srvName, msg)
}

// numGoroutine 并发请求数
func (h Consumer) Run() {
	//init config
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Return.Errors = true
	//config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	if h.resultCallback == nil {
		h.logFatalln("fatal because result callback not setting")
	}
	//消费
	h.logln("connect kafka startup")
	consumerGroup, err := sarama.NewConsumerGroup(h.brokers, h.groupId, config)
	if err != nil {
		h.logln(fmt.Sprintf("NewConsumerGroupFromClient Failed: %s", err.Error()))
		return
	}
	go func() {
	Loop:
		for {
			topics := []string{h.topic}
			// 启动kafka消费组模式，消费的逻辑在上面的 ConsumeClaim 这个方法里
			if err := consumerGroup.Consume(h.cancelCtx, topics, h); err != nil {
				h.logln("Consume Error:" + err.Error())
			}
			//监听退出信号
			select {
			case <-h.cancelCtx.Done():
				break Loop
			default:

			}
		}
		//关闭写通道
		close(h.queue)
	}()
	//开启多个通道消费者
	wg := &sync.WaitGroup{}
	numGo := h.concurrency
	for i := 1; i <= numGo; i++ {
		wg.Add(1)
		go func() {
		Loop:
			for {
				select {
				case data := <-h.queue:
					h.resultCallback(data)
				case <-h.cancelCtx.Done():
					//消费最后一批数据
					for data := range h.queue {
						h.resultCallback(data)
					}
					break Loop
				}
			}
			wg.Done()
		}()
	}
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		h.cancelFunc()
		wg.Wait()
		h.logFatalln(fmt.Sprintf("exit: via signal：%d", runtime.NumGoroutine()))
	}
}

func NewConsumer(srvName string, brokers []string, topic string, groupId string) *Consumer {
	cancelCtx, cancel := context.WithCancel(context.Background())
	consumer := &Consumer{
		srvName:     srvName,
		brokers:     brokers,
		topic:       topic,
		groupId:     groupId,
		queue:       make(chan []byte, 100),
		concurrency: 20,
		cancelCtx:   cancelCtx,
		cancelFunc:  cancel,
	}
	return consumer
}
