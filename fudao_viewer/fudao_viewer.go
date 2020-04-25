package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lilien1010/tx-fudao-crawler/common"
	"github.com/lilien1010/tx-fudao-crawler/model"
	//"io/ioutil"
	"log"

	"github.com/jinzhu/gorm"
	"net/http"
	"runtime"
	//"sync"
)

var (
	infoLog *log.Logger
	Db      *gorm.DB
)

var option common.Options

type ReqParam struct {
	Subject int    `json:"subject" form:"subject"`
	Date    string `json:"date" form:"date"`
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg, err := common.InitLog(&option, &infoLog)

	if err != nil {
		log.Fatal(err)
	}

	Db, err = common.InitMysql(cfg, infoLog, "MYSQL", 100, 2)
	if err != nil {
		log.Fatal(err)
	}

	countInfoModel := model.NewCountHistModel(Db)

	historyModel := model.NewCourseHistoryModel(Db)

	r := gin.Default()

	r.POST("/status", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// 首先加载templates目录下面的所有模版文件，模版文件扩展名随意
	r.LoadHTMLGlob("html/*")

	// 查看数量
	r.GET("/index", func(c *gin.Context) {

		list, err := countInfoModel.QueryCountList()

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "param error " + err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"List": list,
		})
	})

	// 查看数量
	r.GET("/detail", func(c *gin.Context) {

		u := ReqParam{}
		err := c.ShouldBind(&u)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "param error " + err.Error(),
			})
			return
		}

		courseDetail, err := historyModel.QueryDetail(u.Date, u.Subject)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "param error " + err.Error(),
			})
			return
		}
		c.HTML(http.StatusOK, "course.html", gin.H{
			"courseDetail": courseDetail,
		})
	})

	//后期提供HTTP的同步指令
	server_ip := cfg.Section("LOCAL_SERVER").Key("bind_ip").MustString("0.0.0.0")
	listen_port := cfg.Section("LOCAL_SERVER").Key("bind_port").MustString("8021")

	r.Run(server_ip + ":" + listen_port)

}
