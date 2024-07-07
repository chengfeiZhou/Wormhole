package kafka

import (
	"github.com/IBM/sarama"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

var (
	defaultVersion = sarama.V0_11_0_0
)

type OptionFunc func(*Config) error

type Config struct {
	*sarama.Config
	logger           logger.Logger
	producerMsgBatch int // 生产者缓冲队列长度
}

func WithVersion(v sarama.KafkaVersion) OptionFunc {
	return func(c *Config) error {
		c.Config.Version = v
		return nil
	}
}

func WithLogger(log logger.Logger) OptionFunc {
	return func(c *Config) error {
		c.logger = log
		return nil
	}
}
