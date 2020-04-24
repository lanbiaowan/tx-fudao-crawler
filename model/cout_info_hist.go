package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

type CountInfoHist struct {
	Id uint32 `gorm:"column:id;PRIMARY_KEY" json:"id"`

	DateTime    string `gorm:"column:date_time;PRIMARY_KEY" json:"date_time"`
	Subject     int    `gorm:"column:subject;PRIMARY_KEY" json:"subject"`
	CourseCount int    `gorm:"column:course_count;PRIMARY_KEY" json:"course_count"`
	CreateTime  string `gorm:"column:create_time;PRIMARY_KEY" json:"create_time"`
}

type CountHistoryModel struct {
	Db *gorm.DB
}

func (CountHistoryModel) TableName() string {
	return "count_info_hists"
}

func NewCountHistModel(db *gorm.DB) *CountHistoryModel {
	return &CountHistoryModel{
		Db: db,
	}
}

func (a *CountHistoryModel) Insert(data *CountInfoHist) (err error) {

	param := make([]interface{}, 0, 4)
	param = append(param, data.DateTime)
	param = append(param, data.Subject)
	param = append(param, data.CourseCount)
	param = append(param, data.CreateTime)
	stmt := fmt.Sprintf(`INSERT INTO count_info_hists(date_time,subject,course_count,grade,create_time) VALUES(?,?,?,?) %s 
		ON DUPLICATE KEY UPDATE course_count=VALUES(course_count)`)
	return a.Db.Exec(stmt, param...).Error
}

func (a *CountHistoryModel) BulkUpsert(data []CountInfoHist) error {
	str := make([]string, 0, len(data))
	param := make([]interface{}, 0, len(data)*4)
	for _, d := range data {

		str = append(str, "(?,?,?,?)")
		param = append(param, d.DateTime)
		param = append(param, d.Subject)
		param = append(param, d.CourseCount)
		param = append(param, d.CreateTime)

	}
	stmt := fmt.Sprintf(`INSERT INTO count_info_hists(date_time,subject,course_count,grade,create_time) VALUES %s 
		ON DUPLICATE KEY UPDATE course_count=VALUES(course_count)`,
		strings.Join(str, ","))
	return a.Db.Exec(stmt, param...).Error
}

func (a *CountHistoryModel) QueryCountList() (list []CountInfoHist, err error) {

	rawSQL := `SELECT id,date_time,subject,sys_count,course_count,create_time
		FROM count_info_hists
		WHERE order by date_time desc
	`
	rows, err := a.Db.Raw(rawSQL).Rows()
	if err != nil {
		return list, err
	}

	defer rows.Close()

	list = []CountInfoHist{}
	item := CountInfoHist{}
	// 逐条解析
	for rows.Next() {
		a.Db.ScanRows(rows, &item)
		list = append(list, item)
	}

	return list, nil
}
