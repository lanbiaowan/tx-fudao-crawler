package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

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
type CourseCountItem struct {
	DateTime   string `json:"date_time"`
	Chinese    int    `json:"chinese"`
	Math       int    `json:"math"`
	English    int    `json:"english"`
	Geo        int    `json:"geo"`
	Political  int    `json:"political"`
	Bio        int    `json:"bio"`
	Chemistry  int    `json:"chemistry"`
	Physical   int    `json:"physical"`
	Lecture    int    `json:"lecture"`
	Coding     int    `json:"coding"`
	Science    int    `json:"science"`
	CreateTime string `json:"create_time"`
}

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
	stmt := fmt.Sprintf(`INSERT INTO count_info_hists(date_time,subject,course_count,create_time) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE course_count=VALUES(course_count)`)
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

func (a *CountHistoryModel) QueryCountList() (list []*CourseCountItem, err error) {

	rawSQL := `SELECT id,date_time,subject,course_count,create_time
		FROM count_info_hists
		WHERE 1 order by date_time desc
	`
	rows, err := a.Db.Raw(rawSQL).Rows()
	if err != nil {
		return list, err
	}

	defer rows.Close()

	list = []*CourseCountItem{}
	subjectItem := CountInfoHist{}

	uniqueDateList := []string{}

	tempItemMap := map[string]*CourseCountItem{}
	// 逐条解析
	for rows.Next() {
		err := a.Db.ScanRows(rows, &subjectItem)

		if err != nil {
			continue
		}

		val, ok := tempItemMap[subjectItem.DateTime]

		//首次构建
		if !ok {
			item := &CourseCountItem{}
			item.DateTime = subjectItem.DateTime
			ConvertCountInfoIntoItem(&subjectItem, item)
			tempItemMap[item.DateTime] = item
			uniqueDateList = append(uniqueDateList, item.DateTime)
		} else {
			ConvertCountInfoIntoItem(&subjectItem, val)
		}

	}

	for _, v := range uniqueDateList {
		item, ok := tempItemMap[v]
		if ok {
			list = append(list, item)
		}
	}

	return list, nil
}

func ConvertCountInfoIntoItem(info *CountInfoHist, item *CourseCountItem) {

	switch info.Subject {

	case 6001:
		item.Chinese = info.CourseCount
	case 6002:
		item.Math = info.CourseCount
	case 6005:
		item.English = info.CourseCount
	case 6009:
		item.Geo = info.CourseCount
	case 6007:
		item.Political = info.CourseCount
	case 6006:
		item.Bio = info.CourseCount
	case 6003:
		item.Chemistry = info.CourseCount
	case 6004:
		item.Physical = info.CourseCount
	case 6010:
		item.Lecture = info.CourseCount
	case 7057:
		item.Coding = info.CourseCount
	case 7058:
		item.Science = info.CourseCount
	}
}
