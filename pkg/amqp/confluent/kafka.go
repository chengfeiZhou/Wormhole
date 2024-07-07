package confluent

import (
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type OptionFunc func(*Config) error

type Config struct {
	conf *kafka.ConfigMap
	logg logger.Logger
}

// func WithVersion(v sarama.KafkaVersion) OptionFunc {
// 	return func(c *Config) error {
// 		c.Config.Version = v
// 		return nil
// 	}
// }

func WithLogger(log logger.Logger) OptionFunc {
	return func(c *Config) error {
		c.logg = log
		return nil
	}
}
