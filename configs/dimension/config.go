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
	Bind        string `json:"bind"`
	HttpTimeout int    `json:"httpTimeout"`
	// fs
	HandlingPath string `json:"handlingPath"`
	ScanInterval int    `json:"scanInterval"`
	// kafka
	KafkaAddrs     []string `json:"kafkaAddrs"`
	KafkaUser      string   `json:"kafkaUser"`      // 用户名
	KafkaPasswd    string   `json:"kafkaPasswd"`    // 密码
	KafkaMechanism string   `json:"kafkaMechanism"` // 加密算法
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
	fg := flag.NewFlagSetWithEnvPrefix("dimension", envPrefix, flag.ContinueOnError)
	fg.String(flag.DefaultConfigFlagname, "", "配置文件路径(abs)") // 有文件解析文件
	fg.BoolVar(&conf.IsDebug, "isDebug", false, "isDebug")
	fg.IntVar(&conf.ChannelSize, "channelSize", 100, "搬运缓存队列长度")
	// http
	fg.StringVar(&conf.Bind, "bind", "127.0.0.1:8081", "转发目标(ip:port)")
	fg.IntVar(&conf.HttpTimeout, "httpTimeout", 10, "请求超时时间(s)")
	// kafka
	kafkaAddrs := fg.String("kafkaAddrs", "127.0.0.1:9200", "kafka地址(单机:127.0.0.1:9092; 集群:'192.168.100.100:9092,192.168.100.101:9092')")
	fg.StringVar(&conf.KafkaUser, "kafkaUser", "", "kafka鉴权用户名")
	fg.StringVar(&conf.KafkaPasswd, "kafkaPasswd", "", "kafka鉴权密码")
	fg.StringVar(&conf.KafkaMechanism, "kafkaMechanism", "PLAIN", "kafka鉴权的加密算法")
	// FS
	fg.StringVar(&conf.HandlingPath, "handlingPath", files.JoinPath(files.RootAbPathByCaller(), "tmp"),
		"搬运数据中间目录路径")
	fg.IntVar(&conf.ScanInterval, "scanInterval", 5, "搬运目录扫描时间间隔(s)")

	// 执行参数解析
	if err := fg.Parse(args); err != nil {
		return nil, err
	}
	Print(fg)
	// 解析切片值
	conf.KafkaAddrs = strings.Split(*kafkaAddrs, ",")
	return conf, nil
}

// Print 打印解析到的参数
func Print(fg *flag.FlagSet) {
	pFunc := func(f *flag.Flag) {
		// MARK: 辅助打印, 打印配置参数
		fmt.Printf("===> \t %s: %s    Default: %s    Usage: %s\n", f.Name, f.Value.String(), f.DefValue, f.Usage)
	}
	fmt.Printf("**************************************** %s %s ****************************************\n",
		"虫洞", "次元")
	fg.VisitAll(pFunc)
}
