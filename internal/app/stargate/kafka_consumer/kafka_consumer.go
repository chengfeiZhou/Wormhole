package kafkaconsumer

import (
	"context"

	kafka "github.com/chengfeiZhou/Wormhole/pkg/amqp/confluent"

	"github.com/chengfeiZhou/Wormhole/internal/app/stargate"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

const (
	groupID = "wormhole_stargate"
)

// nolint
type Adapter struct {
	Addrs   []string
	Topics  []string
	log     logger.Logger
	msgChan chan<- []byte        // 报文转移通道
	cg      *kafka.ConsumerGroup // kafa的消费者组
}

type OptionFunc func(*Adapter)

// WithLogger 是一个函数选项，用于设置Adapter的日志记录器
// 参数log是logger.Logger类型的日志记录器
// 返回值是一个OptionFunc类型的函数选项，该函数选项用于修改Adapter的日志记录器
func WithLogger(log logger.Logger) OptionFunc {
	return func(a *Adapter) {
		a.log = log
	}
}

// NewAdapter 创建一个新的Adapter实例
// addrs: Kafka的地址列表
// topics: 需要订阅的Kafka主题列表
// msgChan: 接收Kafka消息的通道
// ops: 可选的OptionFunc函数列表，用于配置Adapter
// 返回值：*Adapter，新创建的Adapter实例
func NewAdapter(addrs, topics []string, msgChan chan<- []byte, ops ...OptionFunc) *Adapter {
	ad := &Adapter{
		Addrs:   addrs,
		Topics:  topics,
		log:     logger.DefaultLogger(),
		msgChan: msgChan,
	}
	cg, err := kafka.NewConsumerGroup(addrs, topics, groupID, ad.consumerHandle)
	if err != nil {
		ad.log.Error(logger.ErrorKafkaConsumer, "create consumer group", logger.ErrorField(err))
		return ad
	}
	ad.cg = cg
	for _, op := range ops {
		op(ad)
	}
	return ad
}

// GetName 返回适配器名称，对于此示例，返回的是 "kafka"
func (ad *Adapter) GetName() string {
	return "kafka"
}

// Setup 是Adapter类型的方法，用于设置适配器。
// app参数是传入的stargate.App类型的指针，用于获取日志记录器和配置信息。
// msgChan参数是一个用于发送消息到消息通道的chan<- []byte类型。
// 该方法会返回error类型的结果，如果设置成功则为nil，否则为错误信息。
func (ad *Adapter) Setup(app *stargate.App, msgChan chan<- []byte) error {
	ad.log = app.Logger
	ad.msgChan = msgChan
	conf := app.Config
	kafkaOps := []kafka.OptionFunc{
		kafka.WithLogger(ad.log),
	}
	if conf.KafkaUser != "" && conf.KafkaPasswd != "" {
		kafkaOps = append(kafkaOps, kafka.WitchBaseAuth(conf.KafkaUser, conf.KafkaPasswd, conf.KafkaMechanism))
	}
	cg, err := kafka.NewConsumerGroup(conf.KafkaAddrs, conf.KafkaTopics, groupID, ad.consumerHandle, kafkaOps...)
	if err != nil {
		ad.log.Error(logger.ErrorKafkaConsumer, "create consumer group", logger.ErrorField(err))
		return err
	}
	ad.cg = cg
	ad.Addrs = conf.KafkaAddrs
	ad.Topics = conf.KafkaTopics
	return nil
}

// Help 是一个Adapter类型的方法，用于提供适配器的帮助信息
func (ad *Adapter) Help() {

}

// consumerHandle 是一个Adapter类型的方法，用于处理Kafka消费者接收到的数据
//
// 参数：
//
//	ctx context.Context：上下文对象，用于控制goroutine的生命周期
//	data *kafka.Data：Kafka接收到的数据
//
// 返回值：
//
//	error：处理过程中发生的错误，如果没有错误则返回nil
func (ad *Adapter) consumerHandle(ctx context.Context, data *kafka.Data) error {
	msg := &structs.KafkaMessage{
		Topic:     data.Topic,
		Key:       string(data.Key),
		Value:     data.Value,
		Timestamp: data.FaninTime.UnixMilli(),
	}
	msgB, err := msg.Marshal()
	if err != nil {
		ad.log.Error(logger.ErrorKafkaConsumer, "marshal kafka message", logger.ErrorField(err))
		return err
	}
	ad.msgChan <- msgB
	return nil
}

// Run 执行监听服务
// Run 启动Kafka消费者服务
// ctx: 上下文对象，用于控制服务的启动和停止
// 返回值：
// error: 如果服务启动失败，则返回非零的错误码；否则返回nil
func (ad *Adapter) Run(ctx context.Context) error {
	ad.log.Info("run stargate service for kafka consumer", logger.MakeField("addrs", ad.Addrs),
		logger.MakeField("topics", ad.Topics))
	if ad.cg != nil {
		if err := ad.cg.Run(ctx); err != nil {
			ad.log.Error(logger.ErrorKafkaConsumer, "kafka消费", logger.ErrorField(err))
			return err
		}
	}
	ad.log.Info("stargate service stop for kafka consumer")
	return nil
}
