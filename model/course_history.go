package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

type CourseHistory struct {
	Id uint32 `gorm:"column:id;PRIMARY_KEY" json:"id"`

	DateTime   string `gorm:"column:date_time;PRIMARY_KEY" json:"date_time"`
	Subject    int    `gorm:"column:subject;PRIMARY_KEY" json:"subject"`
	Grade      int    `gorm:"column:grade;PRIMARY_KEY" json:"grade"`
	Price      string `gorm:"column:price;PRIMARY_KEY" json:"price"`
	Title      string `gorm:"column:title;PRIMARY_KEY" json:"title"`
	Teacher    string `gorm:"column:teacher;PRIMARY_KEY" json:"teacher"`
	Detail     string `gorm:"column:detail;PRIMARY_KEY" json:"detail"`
	CreateTime string `gorm:"column:create_time;PRIMARY_KEY" json:"create_time"`
}

type CourseHistoryModel struct {
	Db *gorm.DB
}

func (CourseHistory) TableName() string {
	return "fudao_course_history"
}

func NewCourseHistoryModel(db *gorm.DB) *CourseHistoryModel {
	return &CourseHistoryModel{
		Db: db,
	}
}

func (a *CourseHistoryModel) Insert(order *CourseHistory) error {
	return a.Db.Save(order).Error
}

func (a *CourseHistoryModel) BulkUpsert(data []HistoryData) error {
	str := make([]string, 0, len(data))
	param := make([]interface{}, 0, len(data)*8)
	for _, d := range data {
		str = append(str, "(?,?,?,?,?,?,?,?)")

		param = append(param, d.DateTime)
		param = append(param, d.Subject)
		param = append(param, d.Grade)
		param = append(param, d.Price)

		param = append(param, d.Title)
		param = append(param, d.Teacher)
		param = append(param, d.Detail)
		param = append(param, d.CreateTime)
	}
	stmt := fmt.Sprintf("INSERT INTO fudao_course_history(date_time,subject,price,title,teacher,detail,create_time) VALUES %s ON DUPLICATE KEY UPDATE price=VALUE(price)",
		strings.Join(str, ","))
	return a.Db.Exec(stmt, param...).Error
}
