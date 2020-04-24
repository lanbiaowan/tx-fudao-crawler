package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/lilien1010/tx-fudao-crawler/common"
	"github.com/lilien1010/tx-fudao-crawler/model"

	//"io/ioutil"

	"time"
	//"sync"
	"bytes"
	"errors"
)

var infoLog *log.Logger

const (
	FUDAO_COUSR_START = "window.__initialCgiData={\""
	FUDAO_COUSR_END   = ";</script>"
)

var reqFudaoHeader = map[string]string{
	"referer":    "https://fudao.qq.com/subjec",
	"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36",
}

func init() {
	infoLog = log.New(os.Stdout, "[Info]", log.LstdFlags)
}
func SetLogger(p *log.Logger) {
	infoLog = p
}

//得到课程详情
func GatherCourseDetailByCid(SubjectId int, Cid string) (data *model.HistoryData, err error) {

	reqUrl := fmt.Sprintf("https://fudao.qq.com/pc/course.html?course_id=%s", Cid)

	if err != nil {
		infoLog.Printf("GatherCourseDetailByCid reqUrl=%s,err=%v", reqUrl, err)
		return nil, err
	}

	Content, err := common.HttpGet(reqUrl, 20, reqFudaoHeader)

	if err != nil {
		infoLog.Printf("GatherCourseDetailByCid() simplejson %s,%s new json fail err=%v", reqUrl, string(Content), err)
		return nil, err
	}

	startIndex := bytes.Index(Content, []byte(FUDAO_COUSR_START))

	if startIndex < 0 {
		infoLog.Printf("GatherCourseDetailByCid() simplejson %s fail get start %s", reqUrl, string(FUDAO_COUSR_START))
		return nil, errors.New("end nil")
	}

	endIndex := bytes.Index(Content[startIndex-2:], []byte(FUDAO_COUSR_END))

	if endIndex < 0 {
		infoLog.Printf("GatherCourseDetailByCid() simplejson %s,%s fail get end ", reqUrl, string(Content[startIndex-2:]))
		return nil, errors.New("end nil")
	}

	jsonContent := Content[startIndex-2+len(FUDAO_COUSR_START) : endIndex+startIndex-1]

	curCourse, err := simplejson.NewJson(jsonContent)
	if err != nil {
		infoLog.Printf("GatherCourseDetailByCid() simplejson %s,%s new json fail err=%v", reqUrl, string(jsonContent), err)
		return nil, err
	}

	gradeStr := curCourse.Get("grade").MustString("")
	Grade := 0
	if len(gradeStr) > 0 {
		Grade, _ = strconv.Atoi(gradeStr)
	}

	data = &model.HistoryData{
		CourseId:   curCourse.Get("cid").MustInt(0),
		DateTime:   time.Now().Format("2006-01-02"),
		Subject:    curCourse.Get("subject").MustInt(0),
		Grade:      Grade,
		Price:      curCourse.Get("price").MustInt(0),
		Title:      curCourse.Get("name").MustString(""),
		Teacher:    GetTeacher(curCourse.Get("teacher")),
		Detail:     string(jsonContent),
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	return data, nil
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
