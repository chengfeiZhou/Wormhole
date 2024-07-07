package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/IBM/sarama"
	cvt "github.com/chengfeiZhou/Wormhole/pkg/convert"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
)

const (
	maxMsgBytes         = 200 * 1024 * 1024 // 200MB
	msgBatch            = 10
	batchMsgSubmitBytes = 10 * 1024 * 1024 // 消息达到多大也进行提交
)

type Producer struct {
	conf     *sarama.Config
	producer sarama.SyncProducer // 阻塞的生产者
	logger   logger.Logger
	msgBatch int
}

func ProducerWithBatchSize(size int) OptionFunc {
	return func(c *Config) error {
		c.producerMsgBatch = size
		return nil
	}
}
func ProducerWithMaxMessageBytes(size int) OptionFunc {
	return func(c *Config) error {
		c.Config.Producer.MaxMessageBytes = size
		return nil
	}
}

func NewProducer(addrs []string, ops ...OptionFunc) (*Producer, error) {
	sarama.MaxRequestSize = maxMsgBytes
	conf := &Config{
		Config:           sarama.NewConfig(),
		logger:           logger.DefaultLogger(),
		producerMsgBatch: msgBatch,
	}
	conf.Config.Version = defaultVersion
	// 采用随机而非哈希方法
	conf.Config.Producer.Partitioner = sarama.NewRandomPartitioner
	conf.Config.Producer.Return.Successes = true
	conf.Config.Producer.Return.Errors = true
	conf.Config.Producer.Idempotent = true // 开启幂等性
	conf.Config.Net.MaxOpenRequests = 1    // 开启幂等性后 并发请求数也只能为1
	conf.Config.Producer.RequiredAcks = sarama.WaitForAll
	// conf.Producer.Timeout = time.Duration(5) * time.Minute
	conf.Config.Producer.MaxMessageBytes = maxMsgBytes
	for _, op := range ops {
		if err := op(conf); err != nil {
			return nil, err
		}
	}

	conf.logger.Infof("kafka producer addr: %v", addrs)
	p, err := sarama.NewSyncProducer(addrs, conf.Config)
	if err != nil {
		return nil, err
	}
	prod := &Producer{
		conf:     conf.Config,
		logger:   conf.logger,
		producer: p,
		msgBatch: conf.producerMsgBatch,
	}
	return prod, nil
}

// AsyncPusher 使用通道向kafka发送数据
// TODO: @zcf discuss:这里提供了基于channel的异步生产方式, 但是生产时会产生错误, 这里错误直接打印了;
// 如果后期有需要, 是否可以考虑加一个用来接收error的channel
func (p *Producer) AsyncPusher(ctx context.Context, msgChan <-chan *Msg) {
	arrMsgs := &msgList{msgs: make([]*sarama.ProducerMessage, 0, p.msgBatch)}
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		select {
		case msg := <-msgChan:
			// TODO: @zcf ip资产探测es_insert数据的key是空的
			// if msg == nil || msg.Key == "" || len(msg.Value) == 0 {
			if msg == nil || len(msg.Value) == 0 {
				continue
			}
			arrMsgs.push(msg.makeProducMsg())
			p.logger.Println("  producer async-pusher  topic:", msg.Topic, "len:", len(msg.Value))
			// 达到批次数量和大小都提交
			if arrMsgs.count() >= p.msgBatch || arrMsgs.getMsgLen() >= batchMsgSubmitBytes {
				p.logger.Info("kafka producer send msgs", logger.MakeField("count", arrMsgs.count()),
					logger.MakeField("length", arrMsgs.getMsgLen()))
				if err := p.sendBatch(arrMsgs.getMsgs()); err != nil {
					p.logger.Error(logger.ErrorKafkaProducerSend, "kafka producer send data", logger.ErrorField(err))
				}
				arrMsgs.clear()
			}
		case <-tick.C:
			if err := p.sendBatch(arrMsgs.getMsgs()); err != nil {
				p.logger.Error(logger.ErrorKafkaProducerSend, "kafka producer send data for ticker", logger.ErrorField(err))
			}
			arrMsgs.clear()
		case <-ctx.Done(): // 上下文停止
			p.logger.Info("producer is stopping")
			return
		}
	}
}

func (p *Producer) sendBatch(arrMsgs []*sarama.ProducerMessage) error {
	if len(arrMsgs) < 1 {
		return nil
	}

	if err := p.producer.SendMessages(arrMsgs); err != nil {
		errs, ok := err.(sarama.ProducerErrors)
		if ok {
			resErrs := make(sarama.ProducerErrors, 0)
			for _, item := range errs {
				if _, _, itemErr := p.producer.SendMessage(item.Msg); itemErr != nil {
					p.logger.Infof("singleMsg send error => topic[%s] data: %v; error: %s",
						item.Msg.Topic, item.Msg.Value.Length(), itemErr)
					item.Err = itemErr
					resErrs = append(resErrs, item)
				}
			}
			return resErrs
		}
		return err
	}

	return nil
}

func (p *Producer) Close() {
	if cvt.IsNil(p.producer) {
		if err := p.producer.Close(); err != nil {
			p.logger.Error(logger.ErrorKafkaProducer, "producer close", logger.ErrorField(err))
		}
	}
}

// Msg 定义消息对象
type Msg struct {
	Topic string
	Key   string
	Value []byte
}

// makeProducMsg 构建ProducerMessage
func (m *Msg) makeProducMsg() *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic: m.Topic,
		Key:   sarama.StringEncoder(m.Key),
		Value: sarama.ByteEncoder(m.Value),
	}
}

// msgList 定义待发送数据的对象
type msgList struct {
	msgs   []*sarama.ProducerMessage
	msgLen int
	lock   sync.Mutex
}

func (s *msgList) push(msg *sarama.ProducerMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.msgLen += msg.Value.Length()
	s.msgs = append(s.msgs, msg)
}

func (s *msgList) count() int {
	return len(s.msgs)
}

func (s *msgList) getMsgs() []*sarama.ProducerMessage {
	return s.msgs
}

func (s *msgList) getMsgLen() int {
	return s.msgLen
}

func (s *msgList) clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.msgs = []*sarama.ProducerMessage{}
	s.msgLen = 0
}
