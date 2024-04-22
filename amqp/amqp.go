package amqp

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Client struct {
	amqpConn     *amqp.Connection //链接
	amqpUrl      string           //连接地址
	intervalTime int64            //断线重连检测时间,单位秒,默认5
	amqpClose    chan bool        //关闭通知
}

type Ctrl struct {
	ExchangeName    string //交换机名称
	ExchangeKind    string //交换机类型,直连交换机-direct、主题交换机-topic、头交换机-headers、广播交换机-fanout
	ExchangeDurable bool   //交换机持久化
	QueueName       string //队列名称
	QueueDurable    bool   //队列是否持久化
	IsBindQueue     bool   //是否绑定队列
	RoutingKey      string //key
	ConsumerName    string //消费者名称
}

var AmqpClient = new(Client)

func NewAmqpClient(amqpUrl string, intervalTime int64) *Client {
	if intervalTime <= 0 {
		intervalTime = 5
	}
	AmqpClient = &Client{
		amqpConn:     nil,
		amqpUrl:      amqpUrl,
		intervalTime: intervalTime,
		amqpClose:    make(chan bool),
	}
	return AmqpClient
}

func (a *Client) Keepalive() (err error) {
	var amqpConn *amqp.Connection
	if amqpConn, err = amqp.Dial(a.amqpUrl); err != nil {
		return
	}
	a.amqpConn = amqpConn
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Println("AmqpClient Keepalive Panic:", e)
			}
		}()
		ticker := time.NewTicker(time.Second * time.Duration(a.intervalTime))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if a.amqpConn == nil || a.amqpConn.IsClosed() {
					var (
						aConn   *amqp.Connection
						dialErr error
					)
					if aConn, dialErr = amqp.Dial(a.amqpUrl); dialErr != nil {
						log.Printf("amqp auto connection failed,reason is:%v", dialErr)
						continue
					}
					log.Println("amqp auto connection success.")
					a.amqpConn = aConn
				}
			case <-a.amqpClose:
				break
			}
		}
	}()
	return
}

// Publish 发布消息
func (a *Client) Publish(amqpCtrl Ctrl, bytes []byte) (err error) {
	if bytes == nil || len(bytes) <= 0 {
		err = errors.New("bytes is empty.")
		return
	}
	var amqpChannel *amqp.Channel
	if amqpChannel, err = a.amqpConn.Channel(); err != nil {
		return
	}
	defer func() {
		if e := amqpChannel.Close(); e != nil {
			log.Printf("send message: %s,amqpChannel close failed.reason is :%v", string(bytes), e)
			return
		}
	}()
	//声明交换机
	if err = amqpChannel.ExchangeDeclare(amqpCtrl.ExchangeName, amqpCtrl.ExchangeKind, amqpCtrl.ExchangeDurable, false, false, false, nil); err != nil {
		return
	}
	//声明队列
	/*
	  如果只有一方声明队列,可能会导致下面的情况：
	   a)消费者是无法订阅或者获取不存在的MessageQueue中信息
	   b)消息被Exchange接受以后,如果没有匹配的Queue,则会被丢弃
	   Ps:为了避免上面的问题,所以最好选择两方一起声明
	  	   如果客户端尝试建立一个已经存在的消息队列,RabbitMQ不会做任何事情,并返回客户端建立成功的
	*/
	var queue amqp.Queue
	if queue, err = amqpChannel.QueueDeclare(amqpCtrl.QueueName, amqpCtrl.QueueDurable, false, false, false, nil); err != nil {
		return
	}
	//绑定队列
	if amqpCtrl.IsBindQueue {
		if err = amqpChannel.QueueBind(queue.Name, amqpCtrl.RoutingKey, amqpCtrl.ExchangeName, false, nil); err != nil {
			return
		}
	}
	//发布消息
	return amqpChannel.Publish(amqpCtrl.ExchangeName, amqpCtrl.RoutingKey, false, false, amqp.Publishing{
		Body: bytes,
	})
}

// Subscribe 订阅消息
func (a *Client) Subscribe(amqpCtrl Ctrl, handler func(amqp.Delivery)) (err error) {
	var amqpChannel *amqp.Channel
	if amqpChannel, err = a.amqpConn.Channel(); err != nil {
		return
	}
	defer func() {
		if e := amqpChannel.Close(); e != nil {
			log.Printf("consume message,amqpChannel close failed.reason is :%v", e)
			return
		}
	}()
	//声明交换机
	if err = amqpChannel.ExchangeDeclare(amqpCtrl.ExchangeName, amqpCtrl.ExchangeKind, amqpCtrl.ExchangeDurable, false, false, false, nil); err != nil {
		return
	}
	//声明队列
	var queue amqp.Queue
	if queue, err = amqpChannel.QueueDeclare(amqpCtrl.QueueName, amqpCtrl.QueueDurable, false, false, false, nil); err != nil {
		return
	}
	//绑定队列
	if amqpCtrl.IsBindQueue {
		if err = amqpChannel.QueueBind(queue.Name, amqpCtrl.RoutingKey, amqpCtrl.ExchangeName, false, nil); err != nil {
			return
		}
	}
	var msgChan <-chan amqp.Delivery
	if msgChan, err = amqpChannel.Consume(amqpCtrl.QueueName, amqpCtrl.ConsumerName, true, false, false, false, nil); err != nil {
		return
	}
	for delivery := range msgChan {
		handler(delivery)
	}
	return
}

// Close 资源释放
func (a *Client) Close() {
	if a.amqpConn != nil {
		_ = a.amqpConn.Close()
	}
	close(a.amqpClose)
}
