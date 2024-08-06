package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jmespath/go-jmespath"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/mrdan4es/http_exporter/pkg/config"
)

const MetricsPrefix = "http_exporter"

type collector struct {
	fields []struct {
		desc  *prometheus.Desc
		query string
	}
	client *http.Client
	req    *http.Request
	log    zerolog.Logger
}

func New(ctx context.Context, cfg config.CollectorConfig) (prometheus.Collector, error) {
	log := zerolog.Ctx(ctx).
		With().
		Str("collector", cfg.Name).
		Logger()

	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	switch cfg.Auth.Method {
	case config.AuthMethodBearer:
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Auth.Token))
	case config.AuthMethodXToken:
		req.Header.Set("X-Token", cfg.Auth.Token)
	default:
		return nil, fmt.Errorf("invalid auth method: %s", cfg.Auth.Method)
	}

	c := &collector{
		client: &http.Client{},
		req:    req,
		log:    log,
	}

	for _, f := range cfg.Fields {
		c.fields = append(c.fields, struct {
			desc  *prometheus.Desc
			query string
		}{
			desc: prometheus.NewDesc(fmt.Sprintf("%s_%s_%s", MetricsPrefix, cfg.Name, f.Name),
				f.Description, nil, nil,
			),
			query: f.Query,
		})
	}

	if err := c.check(); err != nil {
		return nil, err
	}

	c.log.Info().Msg("collector initialized")

	return c, nil
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, f := range c.fields {
		ch <- f.desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	data, err := c.get()
	if err != nil {
		c.log.Err(err).Msg("fetch metrics")
		return
	}

	for _, f := range c.fields {
		result, err := jmespath.Search(f.query, data)
		if err != nil {
			c.log.Err(err).Msg("jmespath search")
			continue
		}

		float, ok := result.(float64)
		if !ok {
			c.log.Error().
				Any("result", f).
				Msg("invalid metric type")
			continue
		}

		ch <- prometheus.MustNewConstMetric(f.desc, prometheus.GaugeValue, float)
	}
}

func (c *collector) get() (any, error) {
	r, err := c.client.Do(c.req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *collector) check() error {
	data, err := c.get()
	if err != nil {
		return err
	}

	for _, f := range c.fields {
		result, err := jmespath.Search(f.query, data)
		if err != nil {
			return err
		}

		_, ok := result.(float64)
		if !ok {
			return fmt.Errorf("invalid metric type: %v", result)
		}
	}

	return nil
}
