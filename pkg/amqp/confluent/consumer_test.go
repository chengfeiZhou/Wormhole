package confluent

import (
	"context"
	"fmt"

	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

type ConsumerSuite struct {
	suite.Suite
	// cons *ConsumerGroup
	ctx context.Context
}

func Test_ConsumerSuite(t *testing.T) {
	ctx := context.Background()
	s := &ConsumerSuite{
		ctx: ctx,
	}
	suite.Run(t, s)
}
func (s *ConsumerSuite) BeforeTest(suiteName, testName string) {
	// cons, err := NewConsumerGroup([]string{"172.16.20.30:9092"}, []string{"test-topic"}, "test",
	// 	func(ctx context.Context, data *Data) error {
	// 		fmt.Printf("%+v\n", data)
	// 		return nil
	// 	})
	// assert.NoError(s.T(), err)
}

func (s *ConsumerSuite) Test_ConsumerGroup() {
	convey.Convey("Test_ConsumerGroup", s.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("ConsumerSuite", "ConsumerGroup")
		})
		convey.Convey("consumer_group", func() {
			cons, err := NewConsumerGroup([]string{"10.10.11.130:9092"}, []string{"es_insert"}, "test111",
				func(ctx context.Context, data *Data) error {
					fmt.Printf("%+v\n", data)
					return nil
				})
			convey.So(err, convey.ShouldBeNil)
			go cons.Run(s.ctx)
			<-time.After(30 * time.Second)
		})
	})
}
