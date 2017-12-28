package splunkhec

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	//MockMetrics returns a single event that matches this:
	const validResult = "{\"time\":1257894000,\"event\":\"metric\",\"source\":\"telegraf\",\"host\":\"\",\"fields\":{\"_value\":1,\"metric_name\":\"test1.value\",\"tag1\":\"value1\"}}"

	d := &SplunkHEC{}

	v, _ := json.Marshal(validResult)

	if hecMs, err := buildMetrics(testutil.MockMetrics()[0], d); err == nil {
		b, err := json.Marshal(hecMs)
		if assert.Nil(t, err) {
			assert.Equal(t, v, b)
		}
	}
}
