package model

type HistoryData struct {
	CourseId   int    `json:"course_id"`
	DateTime   string `json:"date_time"`
	Subject    int    `json:"subject"`
	Grade      string `json:"grade"`
	Price      int    `json:"price"` //整数，要除以100
	Title      string `json:"title"`
	Teacher    string `json:"teacher"`
	Detail     string `json:"detail"`
	CreateTime string `json:"create_time"`
}

type CountInfoData struct {
	DateTime    string `json:"date_time"`
	Subject     int    `json:"subject"`
	CourseCount int    `json:"course_count"`
	CreateTime  string `json:"create_time"`
}

type QueueTaskEvent struct {
	Id         uint32        `json:"id"`
	CreateTime uint32        `json:"create_time"` //时间时间戳
	Type       string        `json:"type"`        //type 决定了下面的数据类型 只能是 count_data 或者 history_data
	HisData    []HistoryData `json:"history_data"`
	CountData  CountInfoData `json:"count_data"`
}
