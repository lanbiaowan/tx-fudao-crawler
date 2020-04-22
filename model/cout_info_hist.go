package model

import "github.com/jinzhu/gorm"

type CountInfoHist struct {
	Id uint32 `gorm:"column:id;PRIMARY_KEY" json:"id"`

	DateTime    string `gorm:"column:date_time;PRIMARY_KEY" json:"date_time"`
	Subject     int    `gorm:"column:subject;PRIMARY_KEY" json:"subject"`
	SysCount    int    `gorm:"column:sys_count;PRIMARY_KEY" json:"sys_count"`
	CourseCount int    `gorm:"column:course_count;PRIMARY_KEY" json:"course_count"`
	CreateTime  string `gorm:"column:create_time;PRIMARY_KEY" json:"create_time"`
}

type CountHistoryModel struct {
	Db *gorm.DB
}

func (CountHistoryModel) TableName() string {
	return "fudao_count_hist"
}

func NewCountHistModel(db *gorm.DB) *CountHistoryModel {
	return &CountHistoryModel{
		Db: db,
	}
}

func (a *CountHistoryModel) Insert(order *CountInfoHist) (list []string, err error) {
	return a.Db.Save(order).Error
}

func (a *CountHistoryModel) QueryCountList() error {

	rawSQL := `SELECT date_time,subject
		FROM fudao_count_hist
		WHERE order by date_time desc
	`
	rows, err := a.Db.Raw(rawSQL).Rows()
	if err != nil {
		return list, err
	}

	defer rows.Close()
	// 逐条解析
	for rows.Next() {
		// fmt.Sprintf("SearchNormalUseridByNickname", ...)
		a.Db.ScanRows(rows, &item)
		uidList = append(uidList, item.UserId)
	}

	return list, nil
}
