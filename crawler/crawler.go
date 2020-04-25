package main

import (
	"fmt"
	"strconv"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/lilien1010/tx-fudao-crawler/common"
	"github.com/lilien1010/tx-fudao-crawler/crawler/util"
	"github.com/lilien1010/tx-fudao-crawler/model"

	//"io/ioutil"

	"log"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
	//"sync"
)

var (
	infoLog        *log.Logger
	RedisApi       *common.RedisApi
	RedisQueue     string
	TaskId         uint32
	reqFudaoHeader map[string]string
	//db      *sql.DB
)

var option common.Options

//收集所有课程的数量信息
func Start(SubjectId int, GradeIds []int) (err error) {

	//暂时没考虑分页
	UrlTmp := "https://fudao.qq.com/cgi-proxy/course/discover_subject?client=4&platform=3&version=30&grade=%d&subject=%d&showid=0&page=1&size=10&t=0.4440"

	allCourseIdMap := make(map[int]*simplejson.Json, 128)

	for _, grade := range GradeIds {
		reqUrl := fmt.Sprintf(UrlTmp, grade, SubjectId)

		Content, err := common.HttpGet(reqUrl, 20, reqFudaoHeader)

		if err != nil {
			infoLog.Printf("StartGatherCountInfo reqUrl=%s,err=%v", reqUrl, err)
			continue
		}

		rootObj, err := simplejson.NewJson(Content)
		if err != nil {
			infoLog.Printf("StartGatherCountInfo() simplejson %s,%s new json fail err=%v", reqUrl, string(Content), err)
			continue
		}

		infoLog.Printf("StartGatherCountInfo reqUrl=%s,Content=%s", reqUrl, Content)

		resultCode := rootObj.Get("result").Get("retcode").MustInt(-1)
		if resultCode != 0 {
			infoLog.Printf("StartGatherCountInfo() simplejson %s,%s,result code error", reqUrl, string(Content), err)
			continue
		}

		//收集课程详情到map
		errSpe := GatherCourseDetail(SubjectId, &allCourseIdMap, rootObj.Get("result"))
		if errSpe != nil {
			infoLog.Printf("StartGatherCountInfo() GatherCourseDetail fail %s errSys=%v", reqUrl, errSpe)
			continue
		}

		//把课程入队列
		storeDetail(SubjectId, grade, &allCourseIdMap)
	}

	//任务ID
	newTaskId := atomic.AddUint32(&TaskId, 1)

	countInfo := model.CountInfoData{
		Subject:    SubjectId,
		DateTime:   time.Now().Format("2006-01-02"),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	countInfo.CourseCount = len(allCourseIdMap)

	//把单科的数量信息入库
	listSize, err := RedisApi.PushTask(RedisQueue,
		&model.QueueTaskEvent{
			Id:         newTaskId,
			CreateTime: uint32(time.Now().Unix()),
			Type:       model.TASK_TYPE_COUNT_DATA,
			CountData:  countInfo,
		})

	infoLog.Printf("GatherCountInfo() GatherCountInfo done SubjectId=%d newTaskId=%d listSize=%d countInfo=%v,err=%v",
		SubjectId, newTaskId, listSize, countInfo, err)

	return err
}

//把课程详情再get一次，得到课程详情
func GatherCourseDetail(SubjectId int, allCourseIdMap *map[int]*simplejson.Json, Info *simplejson.Json) (err error) {

	sysInfo := Info.Get("sys_course_pkg_list")

	totalArrayCnt := len(sysInfo.MustArray())

	//系统课程
	for i := 0; i < totalArrayCnt; i++ {
		cidListStr := sysInfo.GetIndex(i).Get("cid_list").MustString("")

		cisList := strings.Split(cidListStr, ",")

		for _, cidStr := range cisList {
			cidId, _ := strconv.Atoi(cidStr)
			if cidId > 0 {
				(*allCourseIdMap)[cidId] = nil
			}
		}

	}

	//专题课 直接取
	courses := Info.Get("spe_course_list").Get("data")

	StoreSysCourseIntoMap(SubjectId, allCourseIdMap, courses)

	return err
}

func storeDetail(SubjectId int, Grade int, allCourseIdMap *map[int]*simplejson.Json) (err error) {

	var courseData *model.HistoryData
	for cid, curCourse := range *allCourseIdMap {

		//专题课，已经有了JSON信息
		if curCourse != nil {
			strMsgBody, _ := curCourse.MarshalJSON()

			courseData = &model.HistoryData{
				CourseId:   curCourse.Get("cid").MustInt(0),
				DateTime:   time.Now().Format("2006-01-02"),
				Subject:    curCourse.Get("subject").MustInt(0),
				Grade:      fmt.Sprintf("%d", Grade),
				Price:      curCourse.Get("af_amount").MustInt(0),
				Title:      curCourse.Get("name").MustString(""),
				Teacher:    util.GetTeacher(curCourse.Get("te_list")),
				Detail:     string(strMsgBody),
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
			}

		} else {
			courseData, err = util.GatherCourseDetailByCid(SubjectId, cid)
		}

		if err != nil {
			continue
		}
		//任务ID
		newTaskId := atomic.AddUint32(&TaskId, 1)

		//把单科的数量信息入库
		listSize, err := RedisApi.PushTask(RedisQueue,
			&model.QueueTaskEvent{
				Id:         newTaskId,
				CreateTime: uint32(time.Now().Unix()),
				Type:       model.TASK_TYPE_HISTORY_DATA,
				HisData:    []model.HistoryData{*courseData},
			})

		infoLog.Printf("storeDetail()  sys_course_pkg_list done SubjectId=%d newTaskId=%d listSize=%d courseData=%#v,err=%v",
			SubjectId, newTaskId, listSize, courseData, err)
	}

	return nil
}

//专题课直接可以获取到，用json 对象装起来
func StoreSysCourseIntoMap(SubjectId int, allCourseIdMap *map[int]*simplejson.Json, CourseListInfo *simplejson.Json) []model.HistoryData {
	coursesCnt := len(CourseListInfo.MustArray())

	HisData := []model.HistoryData{}

	for i := 0; i < coursesCnt; i++ {
		curCourse := CourseListInfo.GetIndex(i)

		Cid := curCourse.Get("cid").MustInt(0)
		if Cid > 0 {
			(*allCourseIdMap)[Cid] = curCourse
		}
	}

	return HisData

}

type ReqParam struct {
	Subject int `json:"subject" form:"subject"`
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg, err := common.InitLog(&option, &infoLog)

	if err != nil {
		log.Fatal(err)
	}

	util.SetLogger(infoLog)

	RedisHost := cfg.Section("REDIS").Key("host").MustString("127.0.0.1:6379")
	RedisPasswd := cfg.Section("REDIS").Key("passwd").MustString("")
	RedisQueue = cfg.Section("REDIS").Key("queue_topic").MustString("fudao_queue")

	RedisApi = common.NewRedisApi(RedisHost, 10000, RedisPasswd)

	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.String(200, "ok")
	})

	reqFudaoHeader = map[string]string{
		"referer":    "https://fudao.qq.com/subjec",
		"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36",
		//"user-agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1",
	}
	/*
		6001：语文
		6002：数学
		6005：英语
		6009：地理
		6007：政治
		6006：生物
		6003：化学
		6004：物理
		6010：讲座
		7057：编程【只有 3 4 5】
		7058：科学
	*/
	allSubject := []int{6001, 6002, 6003, 6004, 6005, 6006, 6007, 6008, 6009, 6010, 7057, 7058}

	//从幼儿园到高中
	allGrade := []int{6001, 6002, 6003, 5001, 5002, 5003, 7001, 7002, 7003, 7004, 7005, 7006, 8001, 8002, 8003}

	//每天load一次
	go func() {
		for {

			time.Sleep(24 * time.Hour)

			//可以考虑并发，但是怕流量受限，暂时这样
			for _, v := range allSubject {
				Start(v, allGrade)
			}
		}
	}()

	//测试入口
	r.GET("/gather", func(c *gin.Context) {
		u := ReqParam{}
		err := c.ShouldBind(&u)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "param error " + err.Error(),
			})
			return
		}

		//测试入口
		if u.Subject == 19901010 {
			//可以考虑并发，但是怕流量受限，暂时这样
			for _, v := range allSubject {
				Start(v, allGrade)
			}
			return
		}

		err = Start(u.Subject, allGrade)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "param error " + err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "ok",
			})
		}

	})

	//后期提供HTTP的同步指令
	server_ip := cfg.Section("LOCAL_SERVER").Key("bind_ip").MustString("0.0.0.0")
	listen_port := cfg.Section("LOCAL_SERVER").Key("bind_port").MustString("8089")

	r.Run(server_ip + ":" + listen_port)

}
