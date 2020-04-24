package util

import (
	"testing"
)

//测试课程详情
func TestGatherCourseDetailByCid(t *testing.T) {

	data, err := GatherCourseDetailByCid(171546, "171546")

	if err != nil {
		t.Fatal("err=", err)
	} else {
		t.Logf("data=%v", *data)
	}
}
