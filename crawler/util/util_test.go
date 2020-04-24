package util

import (
	"testing"
)

//测试课程详情
func TestGatherCourseDetailByCid(t *testing.T) {

	data, err := GatherCourseDetailByCid(132439, "132439")

	if err != nil {
		t.Fatal("err=", err)
	} else {
		t.Logf("data=%v", *data)
	}
}
