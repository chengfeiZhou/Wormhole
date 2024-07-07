package kafkaproducer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/chengfeiZhou/Wormhole/internal/app/dimension"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/amqp/kafka"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

// nolint
type Adapter struct {
	log     logger.Logger
	msgChan <-chan []byte // 报文转移通道
	prod    *kafka.Producer
	Addrs   []string
}

type OptionFunc func(*Adapter)

func NewAdapter(addrs []string, msgChan <-chan []byte, ops ...OptionFunc) *Adapter {
	ad := &Adapter{
		log:     logger.DefaultLogger(),
		msgChan: msgChan,
		Addrs:   addrs,
	}
	producer, err := kafka.NewProducer(addrs)
	if err != nil {
		ad.log.Error(logger.ErrorKafkaProducer, "create kafka producer", logger.ErrorField(err))
	}
	ad.prod = producer
	for _, op := range ops {
		op(ad)
	}
	return ad
}

func (ad *Adapter) GetName() string {
	return "kafka"
}
func (ad *Adapter) Setup(app *dimension.App, msgChan <-chan []byte) error {
	var err error
	ad.log = app.Logger
	ad.msgChan = msgChan

	conf := app.Config
	kafkaOps := []kafka.OptionFunc{
		kafka.WithLogger(ad.log),
	}
	if conf.KafkaUser != "" && conf.KafkaPasswd != "" {
		kafkaOps = append(kafkaOps, kafka.WitchBaseAuth(conf.KafkaUser, conf.KafkaPasswd, sarama.SASLMechanism(conf.KafkaMechanism)))
	}
	ad.Addrs = conf.KafkaAddrs
	ad.prod, err = kafka.NewProducer(conf.KafkaAddrs, kafkaOps...)
	if err != nil {
		ad.log.Error(logger.ErrorKafkaProducer, "create kafka producer", logger.ErrorField(err))
		return err
	}
	return nil
}
func (ad *Adapter) Help() {
	// TODO: go模板语法
	fmt.Println("kafka producer module help")
}
func (ad *Adapter) Transform() structs.TransMessage {
	return structs.TransKafkaMessage
}

// Run 执行监听服务
func (ad *Adapter) Run(ctx context.Context) error {
	ad.log.Info("run service for proxy client for kafka", logger.MakeField("kafka", ad.Addrs))
	asyncSendChan := make(chan *kafka.Msg, cap(ad.msgChan))
	signal := make(chan struct{})
	go func() {
		ad.prod.AsyncPusher(ctx, asyncSendChan)
		signal <- struct{}{}
	}()
	for {
		select {
		case data, ok := <-ad.msgChan:
			if !ok {
				continue
			}
			dataMsg, err := structs.TransKafkaMessage(data)
			if err != nil {
				ad.log.Error(logger.ErrorParam, "搬运数据类型错误", logger.ErrorField(err))
				continue
			}
			dataM := dataMsg.(*structs.KafkaMessage)
			ad.log.Info("kafka数据写入", logger.MakeField(dataM.Topic, dataM.Key), logger.MakeField("timestamp", dataM.Timestamp))
			asyncSendChan <- &kafka.Msg{
				Topic: dataM.Topic,
				Key:   dataM.Key,
				Value: dataM.Value,
			}
		case <-signal:
			ad.log.Info("kafka异步生产者退出")
			return nil
		case <-ctx.Done():
			ad.log.Info("dimension module for kafka execution end")
			return nil
		}
	}
}
