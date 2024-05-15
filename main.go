package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/snappy"
)

type TimeSeries struct {
	Hostgroup string `json:"hostgroup"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Group     string `json:"group"`
}

const MIMIR_URL = "10.10.71.25:9009"

func main() {
	timeSeriesData := []TimeSeries{
		{Hostgroup: "Linux servers", Name: "gong test alert from zabbix", Source: "zabbix_test", Group: "alert"},
		{Hostgroup: "Windows servers", Name: "another test alert", Source: "test_source", Group: "test_group"},
	}
	_, err := Push(timeSeriesData)
	if err != nil {
		fmt.Println(err.Error())
	}

}

// Push the input timeseries to the remote endpoint
func Push(timeseries []TimeSeries) (*http.Response, error) {
	data, err := json.Marshal(timeseries)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	// Create HTTP request
	compressed := snappy.Encode(nil, data)
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/v1/push", MIMIR_URL), bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Set("X-Scope-OrgID", "1")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpClient := &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone()}
	// Execute HTTP request
	res, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return res, nil
}
