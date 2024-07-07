package kafka

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

var (
	SHA256 scram.HashGeneratorFcn = sha256.New
	SHA512 scram.HashGeneratorFcn = sha512.New
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

// WitchBaseAuth 设置基础鉴权 用户名, 密码
// 默认加密协议: SASL/PLAIN
func WitchBaseAuth(user, passwd string, mechanism sarama.SASLMechanism) OptionFunc {
	if len(mechanism) == 0 {
		mechanism = sarama.SASLTypePlaintext
	}
	return func(c *Config) error {
		c.Config.Net.SASL.Enable = true
		c.Config.Net.SASL.User = user
		c.Config.Net.SASL.Password = passwd
		c.Config.Net.SASL.Mechanism = mechanism
		return nil
	}
}
