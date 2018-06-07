package splunkmetric

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/metric"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := SplunkMetricSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)
    expS := []byte(fmt.Sprintf(`{"_time":%d,"_value":91.5,"cpu":"cpu0","metric_name":"cpu.usage_idle"}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := SplunkMetricSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

    expS := []byte(fmt.Sprintf(`{"_time":%d,"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle"}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": "foobar",
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := SplunkMetricSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

    expS := []byte(fmt.Sprintf(`{"_time":%d,"_value":"foobar","cpu":"cpu0","metric_name":"cpu.usage_idle"}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMultiFields(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle":  int64(90),
		"usage_total": 8559615,
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := SplunkMetricSerializer{}
	var buf []byte
	buf, err = s.Serialize(m)
	assert.NoError(t, err)

    expS := []byte(fmt.Sprintf(`{"_time":%d,"_value":90,"cpu":"cpu0","metric_name":"cpu.usage_idle"}`,now.Unix())+"\n"+fmt.Sprintf(`{"_time":%d,"_value":8559615,"cpu":"cpu0","metric_name":"cpu.usage_total"}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}

func TestSerializeMetricWithEscapes(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu tag": "cpu0",
	}
	fields := map[string]interface{}{
		"U,age=Idle": int64(90),
	}
	m, err := metric.New("My CPU", tags, fields, now)
	assert.NoError(t, err)

	s := SplunkMetricSerializer{}
	buf, err := s.Serialize(m)
	assert.NoError(t, err)

    expS := []byte(fmt.Sprintf(`{"_time":%d,"_value":90,"cpu tag":"cpu0","metric_name":"My CPU.U,age=Idle"}`, now.Unix()) + "\n")
	assert.Equal(t, string(expS), string(buf))
}
