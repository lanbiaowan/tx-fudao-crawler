package main

import (
	//"fmt"
	//simplejson "github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/lilien1010/tx-fudao-crawler/common"
	"github.com/lilien1010/tx-fudao-crawler/model"
	//"io/ioutil"
	"encoding/json"
	"log"

	"github.com/jinzhu/gorm"
	"runtime"
	//"sync"
)

var (
	infoLog        *log.Logger
	RedisApi       *common.RedisApi
	reqFudaoHeader map[string]string
	Db             *gorm.DB
)

var option common.Options

//
func fudaoWorker(message *string) (code uint32, err error) {

	task := &model.QueueTaskEvent{}

	json.Unmarshal([]byte(*message), task)

	switch task.Type {

	case "count_data":
		err = HandlerForCountInfo(&task.CountData)
	case "history_data":
		err = HandlerForHisData(task.HisData)
	}

	if err != nil {
		infoLog.Printf("StartFunc() fail task=%v,err=%v",
			task, err)
	}

	return 0, nil
}

//入库到数量表
func HandlerForCountInfo(CountInfo *model.CountInfoData) (err error) {

	dbData := &model.CountInfoHist{
		DateTime:    CountInfo.DateTime,
		Subject:     CountInfo.Subject,
		SysCount:    CountInfo.SysCount,
		CourseCount: CountInfo.CourseCount,
		CreateTime:  CountInfo.CreateTime,
	}

	model := model.NewCountHistModel(Db)

	return model.Insert(dbData)
}

//入库到历史表
func HandlerForHisData(HisData []model.HistoryData) (err error) {

	model := model.NewCourseHistoryModel(Db)

	return model.BulkUpsert(HisData)
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg, err := common.InitLog(&option, &infoLog)

	if err != nil {
		log.Fatal(err)
	}

	RedisHost := cfg.Section("REDIS").Key("host").MustString("127.0.0.1:6379")
	RedisPasswd := cfg.Section("REDIS").Key("passwd").MustString("")
	RedisQueue := cfg.Section("REDIS").Key("queue_topic").MustString("fudao_queue")
	RedisWorkerCnt := cfg.Section("REDIS").Key("worker_count").MustInt(12)

	RedisApi = common.NewRedisApi(RedisHost, 10000, RedisPasswd)

	Db, err = common.InitMySqlGormSection(cfg, infoLog, "MYSQL", 100, 2)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.POST("/status", func(c *gin.Context) {
		c.String(200, "ok")
	})

	//启动消费
	RedisApi.StartWorker(RedisQueue, 15, RedisWorkerCnt, fudaoWorker)

	//后期提供HTTP的同步指令
	server_ip := cfg.Section("LOCAL_SERVER").Key("bind_ip").MustString("0.0.0.0")
	listen_port := cfg.Section("LOCAL_SERVER").Key("bind_port").MustString("8021")

	r.Run(server_ip + ":" + listen_port)

}
