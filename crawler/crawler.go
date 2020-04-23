package main

import (
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/lilien1010/tx-fudao-crawler/common"
	"github.com/lilien1010/tx-fudao-crawler/model"
	//"io/ioutil"
	"bytes"
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

	countInfo := model.CountInfoData{
		Subject:    SubjectId,
		DateTime:   time.Now().Format("2006-01-02"),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, grade := range GradeIds {
		reqUrl := fmt.Sprintf(UrlTmp, grade, SubjectId)

		Content, err := common.HttpGet(reqUrl, 20, reqFudaoHeader)

		if err != nil {
			infoLog.Printf("StartGatherCountInfo reqUrl=%s,err=%v", reqUrl, err)
		} else {

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

			errSys := GatherCountInfo(SubjectId, grade, &countInfo, rootObj.Get("result"))
			if errSys != nil {
				infoLog.Printf("StartGatherCountInfo() GatherCountInfo fail %s errSys=%v", reqUrl, errSys)
				continue
			}

			errSpe := GatherCourseDetail(SubjectId, grade, rootObj.Get("result"))
			if errSpe != nil {
				infoLog.Printf("StartGatherCountInfo() GatherCourseDetail fail %s errSys=%v", reqUrl, errSys)
				continue
			}
		}
	}

	//任务ID
	newTaskId := atomic.AddUint32(&TaskId, 1)

	//把单科的数量信息入库
	listSize, err := RedisApi.PushTask(RedisQueue,
		&model.QueueTaskEvent{
			Id:         newTaskId,
			CreateTime: uint32(time.Now().Unix()),
			Type:       "count_data",
			CountData:  countInfo,
		})

	infoLog.Printf("GatherCountInfo() GatherCountInfo done SubjectId=%d newTaskId=%d listSize=%d countInfo=%v,err=%v",
		SubjectId, newTaskId, listSize, countInfo, err)

	return err
}

//课程课数量统计
func GatherCountInfo(SubjectId int, Grade int, countInfo *model.CountInfoData, Info *simplejson.Json) (err error) {

	sysInfo := Info.Get("sys_course_pkg_list")

	totalArrayCnt := len(sysInfo.MustArray())

	for i := 0; i < totalArrayCnt; i++ {
		cidList := sysInfo.GetIndex(i).Get("cid_list").MustString("")
		//课程包 有多少课程，在这个字段就有多个ID
		countInfo.SysCount += strings.Count(cidList, ",") + 1
	}

	countInfo.CourseCount += Info.Get("spe_course_list").Get("total").MustInt()

	infoLog.Printf("StartGatherCountInfo() GatherCountInfo   SubjectId=%d countInfo=%v", SubjectId, countInfo)

	return err

}

//把课程详情再get一次，得到课程详情
func GatherCourseDetail(SubjectId int, Grade int, Info *simplejson.Json) (err error) {

	sysInfo := Info.Get("sys_course_pkg_list")

	totalArrayCnt := len(sysInfo.MustArray())

	//系统课程
	for i := 0; i < totalArrayCnt; i++ {
		packageId := sysInfo.GetIndex(i).Get("subject_package_id").MustString("")

		reqUrl := fmt.Sprintf("https://fudao.qq.com/grade/%v/subject/%v/subject_system/%s", Grade, SubjectId, packageId)

		Content, err := common.HttpGet(reqUrl, 20, reqFudaoHeader)

		if err != nil {
			infoLog.Printf("GatherCourseDetail reqUrl=%s,err=%v", reqUrl, err)
			continue
		}

		start := "window.__initialState={"
		end := ";</script>"
		index1 := bytes.IndexAny(Content, (start))

		if index1 < 0 {
			infoLog.Printf("GatherCourseDetail reqUrl=%s,can't found data", reqUrl)
			continue
		}

		index2 := bytes.IndexAny(Content[index1:], end)

		infoLog.Println("GatherCourseDetail() simplejson ", reqUrl, index1, index2, string(Content))

		jsonContent := Content[index1+len(start) : index1+index2]
		rootObj, err := simplejson.NewJson(jsonContent)
		if err != nil {
			infoLog.Printf("GatherCourseDetail() simplejson %s,%s new json fail err=%v", reqUrl, string(jsonContent), err)
			continue
		}

		//解析 __initialState 之后的JS 对象。得到 里面的。sysPkgData
		courses := rootObj.Get("sysPkgData").Get("courses")

		HisData := ParseCourseListToHistData(Grade, courses)

		//任务ID
		newTaskId := atomic.AddUint32(&TaskId, 1)

		//把单科的数量信息入库
		listSize, err := RedisApi.PushTask(RedisQueue,
			&model.QueueTaskEvent{
				Id:         newTaskId,
				CreateTime: uint32(time.Now().Unix()),
				Type:       "history_data",
				HisData:    HisData,
			})

		infoLog.Printf("GatherCourseDetail()  sys_course_pkg_list done SubjectId=%d newTaskId=%d listSize=%d countInfo=%v,err=%v",
			SubjectId, newTaskId, listSize, HisData, err)
	}

	//专题课 直接取到
	courses := Info.Get("spe_course_list").Get("data")

	HisData := ParseCourseListToHistData(Grade, courses)

	//任务ID
	newTaskId := atomic.AddUint32(&TaskId, 1)

	//把单科的数量信息入库
	listSize, err := RedisApi.PushTask(RedisQueue,
		&model.QueueTaskEvent{
			Id:         newTaskId,
			CreateTime: uint32(time.Now().Unix()),
			Type:       "history_data",
			HisData:    HisData,
		})

	infoLog.Printf("GatherCourseDetail()  spe_course_list done SubjectId=%d newTaskId=%d listSize=%d countInfo=%v,err=%v",
		SubjectId, newTaskId, listSize, HisData, err)

	return err
}

func ParseCourseListToHistData(Grade int, CourseListInfo *simplejson.Json) []model.HistoryData {
	coursesCnt := len(CourseListInfo.MustArray())

	HisData := []model.HistoryData{}

	for i := 0; i < coursesCnt; i++ {

		curCourse := CourseListInfo.GetIndex(i)

		strMsgBody, _ := curCourse.MarshalJSON()

		data := model.HistoryData{
			CourseId:   curCourse.Get("cid").MustInt(0),
			DateTime:   time.Now().Format("2006-01-02"),
			Subject:    curCourse.Get("subject").MustInt(0),
			Grade:      Grade,
			Price:      curCourse.Get("af_amount").MustInt(0),
			Title:      curCourse.Get("name").MustString(""),
			Teacher:    GetTeacher(curCourse.Get("te_list")),
			Detail:     string(strMsgBody),
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		HisData = append(HisData, data)
	}

	return HisData

}

func GetTeacher(TeInfo *simplejson.Json) string {
	TeachArray := TeInfo.MustArray()
	teach := []string{}
	for i := 0; i < len(TeachArray); i++ {

		tmp := TeInfo.GetIndex(i).Get("name").MustString("")
		teach = append(teach, tmp)

	}

	return strings.Join(teach, ",")
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
		"user-agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1",
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
	allGrade := []int{6001, 6002, 6003, 7001, 7002, 7003, 7004, 7005, 7006, 8001, 8002, 8003}

	//每天load一次
	go func() {
		for {

			time.Sleep(24 * time.Hour)

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
