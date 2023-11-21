package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type balanceCollector struct {
	Balance   *prometheus.Desc
	DaysLeft  *prometheus.Desc
	HoursLeft *prometheus.Desc
	client    *http.Client
	request   *http.Request
}

func NewBalanceCollector(client *http.Client, req *http.Request) *balanceCollector {
	return &balanceCollector{
		Balance: prometheus.NewDesc("regcloud_balance_metric",
			"Balance",
			nil, nil,
		),
		DaysLeft: prometheus.NewDesc("regcloud_days_left_metric",
			"Days left",
			nil, nil,
		),
		HoursLeft: prometheus.NewDesc("regcloud_hours_left_metric",
			"Hours left",
			nil, nil,
		),
		client:  client,
		request: req,
	}
}

func (c *balanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Balance
	ch <- c.DaysLeft
	ch <- c.HoursLeft
}

func (c *balanceCollector) Collect(ch chan<- prometheus.Metric) {
	res, err := c.client.Do(c.request)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer res.Body.Close()

	var response apiResponse

	json.NewDecoder(res.Body).Decode(&response)

	m1 := prometheus.MustNewConstMetric(c.Balance, prometheus.GaugeValue, response.BalanceData.Balance)
	m2 := prometheus.MustNewConstMetric(c.DaysLeft, prometheus.GaugeValue, float64(response.BalanceData.DaysLeft))
	m3 := prometheus.MustNewConstMetric(c.HoursLeft, prometheus.GaugeValue, float64(response.BalanceData.HoursLeft))
	ch <- m1
	ch <- m2
	ch <- m3
}
