package confluent

import (
	"context"
	"strings"
	"time"

	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	offsetCommitLimit = 100
)

type Data struct {
	FaninTime time.Time
	Topic     string
	Key       []byte
	Value     []byte
}

// ConsumerGroup 定义消费者组类
// nolint
type ConsumerGroup struct {
	logger   logger.Logger
	consumer *kafka.Consumer
	topics   []string
	hf       HandleFunc
}

type HandleFunc func(context.Context, *Data) error

// NewConsumerGroup 创建一个消费者实例
// 默认从最旧的开始消费
func NewConsumerGroup(addrs, topics []string, groupID string, handle HandleFunc, ops ...OptionFunc) (*ConsumerGroup, error) {
	conf := &Config{
		conf: &kafka.ConfigMap{
			"bootstrap.servers":         strings.Join(addrs, ","),
			"group.id":                  groupID,
			"auto.offset.reset":         "smallest",
			"api.version.request":       "true",
			"heartbeat.interval.ms":     5000,
			"session.timeout.ms":        120000,
			"max.poll.interval.ms":      120000,
			"fetch.max.bytes":           200 * 1024 * 1024,
			"message.max.bytes":         200 * 1024 * 1024,
			"receive.message.max.bytes": 210 * 1024 * 1024,
			"max.partition.fetch.bytes": 256000,
		},
		logg: logger.DefaultLogger(),
	}
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, err
		}
	}
	consumer, err := kafka.NewConsumer(conf.conf)
	if err != nil {
		return nil, err
	}
	return &ConsumerGroup{
		logger:   conf.logg,
		consumer: consumer,
		topics:   topics,
		hf:       handle,
	}, nil
}

// Run 执行消费动作
func (cg *ConsumerGroup) Run(ctx context.Context) error {
	if err := cg.consumer.SubscribeTopics(cg.topics, nil); err != nil {
		cg.logger.Error(logger.ErrorKafkaConsumer, "create Consumer Group", logger.ErrorField(err))
		return err
	}
	defer cg.consumer.Close()
	cg.logger.Info("run kafka consumer", logger.MakeField("topics", cg.topics))
	msgCount := 0
	var err error
	for {
		ent := cg.consumer.Poll(2000)
		if ent == nil {
			<-time.After(time.Second)
			continue
		}
		switch e := ent.(type) {
		case *kafka.Message:
			err = cg.hf(ctx, &Data{
				FaninTime: e.Timestamp,
				Topic:     *e.TopicPartition.Topic,
				Key:       e.Key,
				Value:     e.Value,
			})
			if err != nil {
				cg.logger.Error(logger.ErrorKafkaConsumer, "Kafka Consumer handle", logger.ErrorField(err))
				continue
			}

			msgCount++
			if msgCount%offsetCommitLimit == 0 {
				// nolint
				offsets, err := cg.consumer.Commit()
				if err != nil {
					cg.logger.Error(logger.ErrorKafkaConsumer, "consumer Commit", logger.ErrorField(err),
						logger.MakeField("TopicPartition", offsets))
				}
				// cg.logger.Info("consumer Commit", logger.MakeField("TopicPartition", offsets))
				continue
			}
			cg.logger.Info("handle message", logger.MakeField("TopicPartition", e.TopicPartition))
		case kafka.PartitionEOF:
			cg.logger.Info("Reached", logger.MakeField("event", e))
		case kafka.Error:
			cg.logger.Error(logger.ErrorKafkaConsumer, "Kafka Consumer data", logger.ErrorField(e))
			return e
		default:
			cg.logger.Debugf("Ignored event: %+v", e)
		}
	}
}
