package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Point struct {
    Fields map[string]float64 `json:"fields"`
    Name   string             `json:"name"`
    Time   time.Time          `json:"time"`
}

type MetricsManager struct {
	meter metric.Meter
	instruments sync.Map
}

func NewMetricsManager() *MetricsManager {
	meter := otel.Meter("server_metrics")
	return &MetricsManager{meter: meter}
}

func ParseMessage(data []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	pointRaw, ok := raw["point"].(map[string]interface{})
	if ok {
		if fields, ok := pointRaw["fields"].(map[string]interface{}); ok {
			for k, v := range fields {
				result[k] = v
			}
		}
		if name, ok := pointRaw["name"]; ok {
			result["point_name"] = name
		}
		if t, ok := pointRaw["time"]; ok {
			result["point_time"] = t
		}
	}

	for k, v := range raw {
		if k == "point" {
			continue
		}
		result[k] = v
	}

	return result, nil
}

func (m *MetricsManager) RecordMetrics(ctx context.Context, data map[string]interface{}) error {
	metricNameRaw, ok := data["point_name"]
	if !ok {
		return fmt.Errorf("missing 'point_name'")
	}
	metricName, ok := metricNameRaw.(string)
	if !ok {
		return fmt.Errorf("'point_name' is not a string")
	}

	instrumentIface, ok := m.instruments.Load(metricName)
	var gauge metric.Float64ObservableGauge
	var err error

	if !ok {
		gauge, err = m.meter.Float64ObservableGauge(
			metricName,
			metric.WithDescription(fmt.Sprintf("Gauge for %s", metricName)),
		)
		if err != nil {
			return err
		}
		m.instruments.Store(metricName, gauge)
	} else {
		gauge = instrumentIface.(metric.Float64ObservableGauge)
	}

	values := make(map[string]float64)
	attrs := make([]attribute.KeyValue, 0)

	for k, v := range data {
		if k == "point_name" || k == "point_time" {
			continue
		}
		if f, ok := toFloat64(v); ok {
			values[k] = f
		} else if s, ok := v.(string); ok {
			attrs = append(attrs, attribute.String(k, s))
		}
	}

	_, err = m.meter.RegisterCallback(
		func(ctx context.Context, observer metric.Observer) error {
			for k, val := range values {
				allAttrs := append(attrs, attribute.String("field", k))
				observer.ObserveFloat64(gauge, val, metric.WithAttributeSet(attribute.NewSet(allAttrs...)))
			}
			return nil
		},
		gauge,
	)

	if err != nil {
		return err
	}

	return nil
}

func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

