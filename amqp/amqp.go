package amqp

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ClientI interface {
	Publish(amqpCtrl Ctrl, bytes []byte) error
	Subscribe(amqpCtrl Ctrl, handler func(amqp.Delivery)) error
}

type amqpClient struct {
	amqpConn    *amqp.Connection
	amqpChannel *amqp.Channel
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

func NewAmqpClient(amqpUrl string, amqpConfig amqp.Config, intervalTime int64) (amqpClientI ClientI, err error) {
	var aClient *amqpClient
	if aClient, err = _AmqpReDial(amqpUrl, amqpConfig); err != nil {
		return
	}
	amqpClientI = aClient
	if intervalTime <= 0 {
		intervalTime = 5
	}
	ticker := time.NewTicker(time.Second * time.Duration(intervalTime))
	go func() {
		for {
			select {
			case <-ticker.C:
				if aClient.amqpConn == nil || aClient.amqpConn.IsClosed() {
					var (
						mqClient *amqpClient
						dialErr  error
					)
					if mqClient, dialErr = _AmqpReDial(amqpUrl, amqpConfig); dialErr != nil {
						log.Println(amqp.ErrClosed.Error())
						continue
					}
					aClient = mqClient
				}
			}
		}
	}()
	return
}

func _AmqpReDial(amqpUrl string, amqpConfig amqp.Config) (*amqpClient, error) {
	var (
		amqpConn    *amqp.Connection
		amqpChannel *amqp.Channel
		err         error
	)

	if amqpConn, err = amqp.DialConfig(amqpUrl, amqpConfig); err != nil {
		return nil, err
	}
	if amqpChannel, err = amqpConn.Channel(); err != nil {
		return nil, err
	}
	return &amqpClient{
		amqpConn:    amqpConn,
		amqpChannel: amqpChannel,
	}, nil
}

// Publish 发布消息
func (a *amqpClient) Publish(amqpCtrl Ctrl, bytes []byte) (err error) {
	if a.amqpConn == nil || a.amqpConn.IsClosed() {
		err = errors.New(amqp.ErrClosed.Error())
		return
	}
	if bytes == nil || len(bytes) <= 0 {
		err = errors.New("bytes is empty!.")
		return
	}
	//声明交换机
	if err = a.amqpChannel.ExchangeDeclare(amqpCtrl.ExchangeName, amqpCtrl.ExchangeKind, amqpCtrl.ExchangeDurable, false, false, false, nil); err != nil {
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
	if queue, err = a.amqpChannel.QueueDeclare(amqpCtrl.QueueName, amqpCtrl.QueueDurable, false, false, false, nil); err != nil {
		return
	}
	//绑定队列
	if amqpCtrl.IsBindQueue {
		if err = a.amqpChannel.QueueBind(queue.Name, amqpCtrl.RoutingKey, amqpCtrl.ExchangeName, false, nil); err != nil {
			return
		}
	}
	//发布消息
	return a.amqpChannel.Publish(amqpCtrl.ExchangeName, amqpCtrl.RoutingKey, false, false, amqp.Publishing{
		Body: bytes,
	})
}

// Subscribe 订阅消息
func (a *amqpClient) Subscribe(amqpCtrl Ctrl, handler func(amqp.Delivery)) (err error) {
	if a.amqpConn == nil || a.amqpConn.IsClosed() {
		err = errors.New(amqp.ErrClosed.Error())
		return
	}
	if _, err = a.amqpChannel.QueueDeclare(amqpCtrl.QueueName, amqpCtrl.QueueDurable, false, false, false, nil); err != nil {
		return
	}
	var msgChan <-chan amqp.Delivery
	if msgChan, err = a.amqpChannel.Consume(amqpCtrl.QueueName, amqpCtrl.ConsumerName, true, false, false, false, nil); err != nil {
		return
	}
	for delivery := range msgChan {
		handler(delivery)
	}
	return
}

// Close 资源释放
func (a *amqpClient) Close() {
	if a.amqpChannel != nil {
		_ = a.amqpChannel.Close()
	}
	if a.amqpConn != nil {
		_ = a.amqpConn.Close()
	}
}
