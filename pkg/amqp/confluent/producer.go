package confluent

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	p   *kafka.Producer
	log logger.Logger
}

func NewProducer(addrs []string, ops ...OptionFunc) (*Producer, error) {
	conf := &Config{
		conf: &kafka.ConfigMap{
			"bootstrap.servers": strings.Join(addrs, ","),
			"client.id":         fmt.Sprintf("wormhole-%d", rand.Int()),
			"acks":              "all",
		},
		logg: logger.DefaultLogger(),
	}
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, err
		}
	}
	conf.logg.Println("  confluent producer addr:", addrs)
	p, err := kafka.NewProducer(conf.conf)
	if err != nil {
		return nil, err
	}
	pro := &Producer{
		p:   p,
		log: conf.logg,
	}

	return pro, nil
}

// AsyncPusher 使用通道向kafka发送数据
// TODO: @zcf discuss:这里提供了基于channel的异步生产方式, 但是生产时会产生错误, 这里错误直接打印了;
// 如果后期有需要, 是否可以考虑加一个用来接收error的channel
func (p *Producer) AsyncPusher(ctx context.Context, msgChan <-chan *Msg) {
}

// PushMsgs 批量的向kafka的同一个topic发送数据
// PS: @zcf这里不对推送数量数量做限制和校验
func (p *Producer) BatchMsgsPush(ctx context.Context, msgs []*Msg) error {
	return nil
}

// SinglePushMsg 向kafka的一个topic发送一个数据
func (p *Producer) SingleMsgPush(ctx context.Context, msg *Msg) error {
	deliveryChan := make(chan kafka.Event, 3)
	defer close(deliveryChan)
	if err := p.p.Produce(msg.makeProducMsg(), deliveryChan); err != nil {
		return err
	}
	ent := <-deliveryChan
	m := ent.(*kafka.Message)
	if err := m.TopicPartition.Error; err != nil {
		p.log.Error(logger.ErrorKafkaProducerSend, "Delivery failed", logger.ErrorField(err))
		return err
	}
	p.log.Info("Delivered message", logger.MakeField("topic", *m.TopicPartition.Topic),
		logger.MakeField("Partition", m.TopicPartition.Partition), logger.MakeField("Offset", m.TopicPartition.Offset))
	return nil
}

func (p *Producer) Close() {

}

// Msg 定义消息对象
type Msg struct {
	Topic string
	Key   string
	Value []byte
}

// makeProducMsg 构建ProducerMessage
func (m *Msg) makeProducMsg() *kafka.Message {
	return &kafka.Message{
		// Timestamp      time.Time
		// TimestampType  TimestampType
		// Opaque         interface{}
		// Headers        []Header
		TopicPartition: kafka.TopicPartition{Topic: &m.Topic, Partition: kafka.PartitionAny},
		Key:            []byte(m.Key),
		Value:          m.Value,
	}
}
