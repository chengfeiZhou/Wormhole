package confluent

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
	pro *Producer
	ctx context.Context
}

func Test_MainSuite(t *testing.T) {
	ctx := context.Background()
	s := &MainSuite{
		ctx: ctx,
	}
	suite.Run(t, s)
}
func (s *MainSuite) BeforeTest(suiteName, testName string) {
	p, err := NewProducer([]string{"172.16.20.30:9092"})
	assert.NoError(s.T(), err)
	s.pro = p
}

func (s *MainSuite) Test_SingleMsgPush() {
	convey.Convey("Test_SingleMsgPush", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "SingleMsgPush")
		})
		convey.Convey("single_msg_push", func() {
			msg := &Msg{
				Topic: "test-topic",
				Key:   "1234567890",
				Value: []byte("hello world"),
			}
			err := s.pro.SingleMsgPush(s.ctx, msg)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func (s *MainSuite) Test_BatchMsgPush() {
	convey.Convey("Test_SingleMsgPush", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "SingleMsgPush")
		})
		convey.Convey("single_msg_push", func() {
			for i := 0; i < 10; i++ {
				msg := &Msg{
					Topic: "test-topic",
					Key:   gofakeit.StreetName(),
					Value: []byte(gofakeit.StreetName()),
				}
				err := s.pro.SingleMsgPush(s.ctx, msg)
				convey.So(err, convey.ShouldBeNil)
			}
		})
	})
}
