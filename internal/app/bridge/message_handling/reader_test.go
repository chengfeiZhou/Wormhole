package messagehandling

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/chengfeiZhou/Wormhole/internal/pkg/structs"
	"github.com/chengfeiZhou/Wormhole/pkg/logger"
	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

func TestRenameReadyFile(t *testing.T) {
	tmpPath := files.JoinPath(files.RootAbPathByCaller(), "internal/app/bridge/file_handling/tmp")
	convey.Convey("rename ready file", t, func() {
		convey.Reset(func() {
			_ = os.RemoveAll(tmpPath)
		})
		convey.Convey("simple file", func() {
			// 先创建tmp
			err := files.IsNotExistMkDir(tmpPath)
			convey.So(err, convey.ShouldBeEmpty)

			f, err := os.Create(files.JoinPath(tmpPath, "test.hsxa"))
			convey.So(err, convey.ShouldBeEmpty)
			newname, err := renameReadyFile(f.Name())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(newname, convey.ShouldEqual, files.JoinPath(tmpPath, "test.hsxa_bak"))
			convey.So(files.CheckExist(newname), convey.ShouldBeTrue)
		})
	})

	convey.Convey("restore file to ready", t, func() {
		convey.Reset(func() {
			_ = os.RemoveAll(tmpPath)
		})
		convey.Convey("simple file", func() {
			// 先创建tmp
			err := files.IsNotExistMkDir(tmpPath)
			convey.So(err, convey.ShouldBeEmpty)

			f, err := os.Create(files.JoinPath(tmpPath, "test.hsxa_bak"))
			convey.So(err, convey.ShouldBeEmpty)
			oldname, err := restoreFileToReady(f.Name())
			convey.So(err, convey.ShouldBeEmpty)
			convey.So(oldname, convey.ShouldEqual, files.JoinPath(tmpPath, "test.hsxa"))
			convey.So(files.CheckExist(oldname), convey.ShouldBeTrue)
		})
	})
}

type MainSuite struct {
	suite.Suite
	reader  *Reader
	msgChan chan<- []byte
}

func Test_MainSuite(t *testing.T) {
	s := &MainSuite{}
	suite.Run(t, s)
}

func transformTestfunc([]byte) (structs.Message, error) {
	return nil, nil
}

func (s *MainSuite) BeforeTest(suiteName, testName string) {
	s.msgChan = make(chan<- []byte, 3)
	s.reader = NewReader(
		s.msgChan,
		WithLoggerToRead(logger.NopLogger()),
		WithHandlingPathToRead(os.TempDir()),
		WithScanIntervalToRead(100*time.Millisecond), // 测试, 用100ms
	)
}
func (s *MainSuite) TestReaderRun() {
	convey.Convey("Test_ReaderRun", s.Suite.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "TestReaderRun")
		})
		convey.Convey("handle ok", func() {
			defer gomonkey.ApplyFunc(s.reader.handling, func() error {
				return nil
			}).Reset()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			err := s.reader.Run(ctx)
			convey.So(err, convey.ShouldBeEmpty)
		})
		convey.Convey("handle error", func() {
			defer gomonkey.ApplyFunc(s.reader.handling, func() error {
				return errors.New("throwing an exception")
			}).Reset()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			err := s.reader.Run(ctx)
			convey.So(err, convey.ShouldBeEmpty)
		})
	})
}

func (s *MainSuite) TestReaderHandling() {
	convey.Convey("Test_ReaderHandling", s.Suite.T(), func() {
		convey.Reset(func() {
			s.BeforeTest("MainSuite", "TestReaderHandling")
		})

		convey.Convey("handle ok", func() {
		})
	})
}
