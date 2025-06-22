package common

import (
	"testing"
	"time"
)

var _ = GetLogger("test")

func TestSleepUntilAccurate(t *testing.T) {
	targetTime := time.Now().Add(300 * time.Second)
	err := SleepUntilAccurate(targetTime)
	if err != nil {
		t.Fatalf("SleepUntilAccurate 执行失败: %v", err)
	}
	ntpAfter := GetAccurateTime()
	drift := ntpAfter.Sub(targetTime)
	allowedDrift := 50 * time.Millisecond
	if drift < -allowedDrift || drift > allowedDrift {
		t.Errorf("时间误差过大: %v (允许范围 ±%v)", drift, allowedDrift)
	} else {
		t.Logf("测试通过，时间误差: %v", drift)
	}
}
