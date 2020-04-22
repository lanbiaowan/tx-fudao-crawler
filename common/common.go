package common

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	//"go.uber.org/zap"
	//"encoding/json"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"net/http"
	//"time"
)

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`

	// Example of optional value
	ServerConf string `short:"c" long:"conf" description:"Server Config" optional:"no"`

	TestCount int `short:"n" long:"testcount" description:"test cout"`
}

func InitLog(options *Options, infoLog **log.Logger) (conf *ini.File, err error) {

	var parser = flags.NewParser(options, flags.Default)

	if _, err := parser.Parse(); err != nil {
		log.Fatalln("InitLogAndOption() main parse cmd line failed!")
	}

	if options.ServerConf == "" {
		log.Fatalln("InitLogAndOption() main Must input config file name")
	}
	cfg, err := ini.Load([]byte(""), options.ServerConf)
	if err != nil {
		log.Fatalln("main load config file=%s failed", options.ServerConf)
		return nil, nil
	}

	cfg.BlockMode = false

	fileName := cfg.Section("LOG").Key("path").MustString("")

	if fileName != "" {
		logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("open file error !,fileName=", fileName)
			return nil, nil
		}
		*infoLog = log.New(logFile, "[Info]", log.LstdFlags)
	} else {
		*infoLog = log.New(os.Stdout, "[Info]", log.LstdFlags)
	}

	(*infoLog).SetFlags((*infoLog).Flags() | log.LstdFlags)

	return cfg, nil
}

//gorm db
func InitMySqlGormSection(cfg *ini.File, infoLog *log.Logger, secName string, OpenCount int, OdleCount int) (*gorm.DB, error) {

	infoLog.Printf("InitMySqlGormSection() name=%s,Idle=%d,Open=%d start", secName, OdleCount, OpenCount)

	// init mysql
	mysqlHost := cfg.Section(secName).Key("mysql_host").MustString("127.0.0.1")
	mysqlUser := cfg.Section(secName).Key("mysql_user").MustString("IMServer")
	mysqlPasswd := cfg.Section(secName).Key("mysql_passwd").MustString("hello")
	mysqlDbName := cfg.Section(secName).Key("mysql_db").MustString("HT_IMDB")
	mysqlPort := cfg.Section(secName).Key("mysql_port").MustString("3306")

	mydb, err := gorm.Open("mysql", mysqlUser+":"+mysqlPasswd+"@"+"tcp("+mysqlHost+":"+mysqlPort+")/"+mysqlDbName+"?charset=utf8mb4&timeout=10s")
	if err != nil {
		infoLog.Println("open ", secName, " failed", time.Now())
		return nil, err
	}

	if err := mydb.DB().Ping(); err != nil {
		infoLog.Fatalln(secName, " ping err", err)
		return nil, err
	}

	mydb.DB().SetMaxIdleConns(OdleCount)

	mydb.DB().SetMaxOpenConns(OpenCount)

	infoLog.Printf("InitMySqlGormSection() name=%s,Idle=%d,Open=%d end", secName, OdleCount, OpenCount)

	return mydb, nil
}

func HttpGet(url string, timeout int, headers map[string]string) ([]byte, error) {

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Second*time.Duration(timeout))
				if err != nil {
					return nil, err
				}
				c.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second)) //设置发送接收数据超时
				return c, nil
			},
		},
		Timeout: time.Second * time.Duration(timeout),
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return (body), nil
}
