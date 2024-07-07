package confluent

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// WitchBaseAuth 设置基础鉴权 用户名, 密码
// 默认加密协议: SASL/PLAIN
func WitchBaseAuth(user, passwd, securityProtocol string) OptionFunc {
	return func(cfg *Config) error {
		switch securityProtocol {
		case "PLAINTEXT":
			_ = cfg.conf.SetKey("security.protocol", "plaintext")
		// case "SASL_SSL":
		// 	cfg.conf.SetKey("security.protocol", "sasl_ssl")
		// 	cfg.conf.SetKey("ssl.ca.location", "./conf/mix-4096-ca-cert")
		// 	cfg.conf.SetKey("sasl.username", cfg.SaslUsername)
		// 	cfg.conf.SetKey("sasl.password", cfg.SaslPassword)
		// 	cfg.conf.SetKey("sasl.mechanism", cfg.SaslMechanism)
		case "SASL_PLAINTEXT", "PLAIN":
			_ = cfg.conf.SetKey("security.protocol", "sasl_plaintext")
			_ = cfg.conf.SetKey("sasl.username", user)
			_ = cfg.conf.SetKey("sasl.password", passwd)
			_ = cfg.conf.SetKey("sasl.mechanism", "PLAIN")
		default:
			return kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol", true)
		}
		return nil
	}
}
