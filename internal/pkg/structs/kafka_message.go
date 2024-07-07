package structs

import (
	"encoding/base64"
	"encoding/json"
)

// nolint
type KafkaMessage struct {
	Topic     string `json:"topic"`
	Key       string `json:"key"`
	Value     []byte `json:"-"`
	ValueB64  string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// Unmarshal []]byte => 报文结构
func (msg *KafkaMessage) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, msg); err != nil {
		return err
	}
	v, err := base64.StdEncoding.DecodeString(msg.ValueB64)
	if err != nil {
		return err
	}
	msg.Value = v
	return nil
}

// Marshal 报文结构 => []byte
func (msg *KafkaMessage) Marshal() ([]byte, error) {
	msg.ValueB64 = base64.StdEncoding.EncodeToString(msg.Value)
	return json.Marshal(msg)
}

func TransKafkaMessage(line []byte) (Message, error) {
	data := new(KafkaMessage)
	err := data.Unmarshal(line)
	return data, err
}
