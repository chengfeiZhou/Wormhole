package config

import (
	"fmt"
	"strings"

	"github.com/chengfeiZhou/Wormhole/pkg/storage/files"
	"github.com/namsral/flag"
)

// nolint
type Config struct {
	IsDebug     bool `json:"isDebug"`
	ChannelSize int  `json:"channelSize"`
	// http
	Listen string `json:"bind"`
	// kafka
	KafkaAddrs     []string `json:"kafkaAddrs"`
	KafkaTopics    []string `json:"kafkaTopics"`
	KafkaUser      string   `json:"kafkaUser"`      // 用户名
	KafkaPasswd    string   `json:"kafkaPasswd"`    // 密码
	KafkaMechanism string   `json:"kafkaMechanism"` // 加密算法

	// file
	HandlingPath string `json:"handlingPath"`
	WriterTicker int    `json:"writerTicker"` // 文件写入间隔
	FileMaxSize  int64  `json:"fileMaxSize"`  // 单文件最大(byte)
}

/*
Order of precedence:
	- Command line options
	- Environment variables
	- Configuration file
	- Default values
*/
/*
参数定义:
|Type	|Flag			|Prefix Environment		|File			|
|-------|---------------|-----------------------|---------------|
|int	|-age 2			|WORMHOLE_AGE=2			|age 2			|
|bool	|-female		|WORMHOLE_FEMALE=true	|female true	|
|float	|-length 175.5	|WORMHOLE_LENGTH=175.5	|length 175.5	|
|string	|-name Gloria	|WORMHOLE_NAME=Gloria	|name Gloria	|
*/

const (
	envPrefix = "WORMHOLE"
)

// PS: 我不想直接使用`init()` -- zcf

func InitConfig(args []string) (*Config, error) {
	conf := new(Config)
	fg := flag.NewFlagSetWithEnvPrefix("stargate", envPrefix, flag.ContinueOnError)
	fg.String(flag.DefaultConfigFlagname, "", "配置文件路径(abs)") // 有文件解析文件
	fg.BoolVar(&conf.IsDebug, "isDebug", false, "isDebug")
	fg.IntVar(&conf.ChannelSize, "channelSize", 100, "搬运缓存队列长度")
	// http
	fg.StringVar(&conf.Listen, "listen", "0.0.0.0:8080", "对外提供服务")
	// kafka
	kafkaAddrs := fg.String("kafkaAddrs", "127.0.0.1:9092", "kafka地址(单机:127.0.0.1:9092; 集群:'192.168.100.100:9092,192.168.100.101:9092')")
	kafkaTopics := fg.String("kafkaTopics", "", "订阅topic('topic1,topic2')")
	fg.StringVar(&conf.KafkaUser, "kafkaUser", "", "kafka鉴权用户名")
	fg.StringVar(&conf.KafkaPasswd, "kafkaPasswd", "", "kafka鉴权密码")
	fg.StringVar(&conf.KafkaMechanism, "kafkaMechanism", "PLAIN", "kafka鉴权的加密算法")

	// file
	fg.StringVar(&conf.HandlingPath, "handlingPath", files.JoinPath(files.RootAbPathByCaller(), "tmp"), "搬运数据中间目录路径")
	fg.IntVar(&conf.WriterTicker, "writerTick", 1, "写文件间隔(s)")
	fg.Int64Var(&conf.FileMaxSize, "fileMaxSize", 10<<20, "单文件最大(byte)")

	// 执行参数解析
	if err := fg.Parse(args); err != nil {
		return nil, err
	}
	Print(fg)
	// 解析切片值
	conf.KafkaAddrs = strings.Split(*kafkaAddrs, ",")
	conf.KafkaTopics = strings.Split(*kafkaTopics, ",")
	return conf, nil
}

// Print 打印解析到的参数
func Print(fg *flag.FlagSet) {
	pFunc := func(f *flag.Flag) {
		// MARK: 辅助打印, 打印配置参数
		fmt.Printf("===> \t %s: %s    Default: %s    Usage: %s\n", f.Name, f.Value.String(), f.DefValue, f.Usage)
	}
	fmt.Printf("**************************************** %s %s ****************************************\n",
		"虫洞", "星门")
	fg.VisitAll(pFunc)
}
