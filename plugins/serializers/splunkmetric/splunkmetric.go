package splunkmetric

import (
    ejson "encoding/json"
    "time"

    "github.com/influxdata/telegraf"
)

type SplunkMetricSerializer struct {
    TimestampUnits time.Duration
}

func (s *SplunkMetricSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
    //m := make(map[string]interface{})
    units_nanoseconds := s.TimestampUnits.Nanoseconds()
    // if the units passed in were less than or equal to zero,
    // then serialize the timestamp in seconds (the default)
    if units_nanoseconds <= 0 {
        units_nanoseconds = 1000000000
    }

    var serialized string

    /* Splunk supports one metric per line and has the following required names:
    ** metric_name: The name of the metric
    ** _value:      The value for the metric
    ** _time:       The timestamp for the metric
    ** All other index fields become deminsions.
    */
    for k,v := range metric.Fields() {
        m := make(map[string]interface{})

        // Break tags out into key(n)=value(t) pairs
        for n,t := range metric.Tags() {
            m[n] = t
        }
        m["metric_name"] = metric.Name()+"."+k
        m["_value"] = v
        m["_time"] = metric.Time().UnixNano() / units_nanoseconds
        metricJson, err := ejson.Marshal(m)
        if err != nil {
            return []byte{}, err
        }
        metricJson = append(metricJson, '\n')
        serialized = serialized+string(metricJson)
    }

    return []byte(serialized), nil
}

