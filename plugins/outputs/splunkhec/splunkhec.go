package splunkhec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type SplunkHEC struct {
	Token  string
	Url    string
	Index  string
	Source string

	Timeout internal.Duration
	client  *http.Client
}

var sampleConfig = `
  ## Splunk HEC Token (also used for Data Channel ID)
  token = "my-secret-key" # required.

  ## Splunk HEC endpoint
  url = "https://localhost:8088/services/collector" # required.

  ## Splunk Index: Must be a metrics index, must be allowed by above token
  # index = "telegraf"

  ## Source: Set the 'source' on the events (defaults to: telegraf)
  # source = "telegraf"

  ## Connection timeout.
  # timeout = "5s"
`

const redactedAPIKey = "****************"

/* Splunk HEC Structs */
type HECTimeSeries struct {
	Time   float64                `json:"time"`
	Event  string                 `json:"event"`
	Index  string                 `json:"index,omitempty"`
	Source string                 `json:"source,omitempty"`
	Host   string                 `json:"host"`
	Fields map[string]interface{} `json:"fields"`
}

func NewSplunkHEC() *SplunkHEC {
	return &SplunkHEC{}
}

func (d *SplunkHEC) Connect() error {
	if d.Token == "" {
		return fmt.Errorf("token is a required field for Splunk HEC output")
	}
	if d.Url == "" {
		return fmt.Errorf("url is a required field for Splunk HEC output")
	}

	d.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: d.Timeout.Duration,
	}
	return nil
}

func (d *SplunkHEC) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}
	var hecPostData string

	for _, m := range metrics {
		if hecMs, err := buildMetrics(m, d); err == nil {
			hecPostData = hecPostData + string(hecMs)
		} else {
			log.Printf("I! unable to build Metric for %s, skipping\n", m.Name())
		}
	}

	req, err := http.NewRequest("POST", d.Url, strings.NewReader(hecPostData))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", strings.Replace(err.Error(), d.Token, redactedAPIKey, -1))
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk "+d.Token)
	// Add the Request-Channel header incase Indexer Acknowledgment is enabled.
	req.Header.Add("X-Splunk-Request-Channel", d.Token)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", strings.Replace(err.Error(), d.Token, redactedAPIKey, -1))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("received bad status code, %d [%s]\n", resp.StatusCode, body)
	}

	return nil
}

func (d *SplunkHEC) SampleConfig() string {
	return sampleConfig
}

func (d *SplunkHEC) Description() string {
	return "Configuration for Splunk HEC to send metrics to.\nDoes not make use of Indexer Acknowledgement"
}

func buildMetrics(m telegraf.Metric, d *SplunkHEC) (metricGroup string, err error) {
	for k, v := range m.Fields() {
		if !verifyValue(v) {
			continue
		}
		obj := map[string]interface{}{}
		obj["metric_name"] = m.Name() + "." + k
		obj["_value"] = v

		dataGroup := HECTimeSeries{
			Time:   float64(m.Time().UnixNano() / 1000000000),
			Event:  "metric",
			Fields: obj,
		}

		if d.Source == "" {
			dataGroup.Source = "telegraf"
		} else {
			dataGroup.Source = d.Source
		}
		dataGroup.Index = d.Index

		// Need to get Host from m.Tags()
		buildHecTags(m, &dataGroup)
		b, err := json.Marshal(dataGroup)
		if err != nil {
			return metricGroup, err
		}
		metricGroup = metricGroup + string(b)
	}

	return metricGroup, nil
}

func buildHecTags(m telegraf.Metric, tsData *HECTimeSeries) {
	/*
	 ** Iterate tags and copy them into fields{}
	 ** Check for host in m.Tags() and set in tsData.Host
	 */
	for k, v := range m.Tags() {
		if k == "host" {
			tsData.Host = v
		} else {
			tsData.Fields[k] = v
		}
	}
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}

func (d *SplunkHEC) Close() error {
	return nil
}

func init() {
	outputs.Add("splunkhec", func() telegraf.Output {
		return NewSplunkHEC()
	})
}
