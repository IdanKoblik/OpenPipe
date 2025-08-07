package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)


type MetricsManager struct {
	meter metric.Meter
	instruments sync.Map
}

func NewMetricsManager() *MetricsManager {
	meter := otel.Meter("server_metrics")
	return &MetricsManager{meter: meter}
}

func Flatten(data map[string]interface{}, prefix string, result map[string]interface{}) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			Flatten(val, fullKey, result)
		default:
			result[fullKey] = val
		}
	}
}

func ParseMessage(data []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	Flatten(raw, "", result)
	return result, nil
}

func (m *MetricsManager) RecordMetrics(ctx context.Context, data map[string]interface{}) error {
	for key, rawVal := range data {
		 if val, ok := toFloat64(rawVal); ok {
			instrumentIface, ok := m.instruments.Load(key)
			var gauge metric.Float64ObservableGauge
			var err error

			if !ok {
				gauge, err = m.meter.Float64ObservableGauge(key, metric.WithDescription(fmt.Sprintf("Gauge for %s", key)),)
				if err != nil {
					return err
				}

				m.instruments.Store(key, gauge)
			} else {
				gauge = instrumentIface.(metric.Float64ObservableGauge)
			}

			_, err = m.meter.RegisterCallback(
             func(ctx context.Context, observer metric.Observer) error {
					 observer.ObserveFloat64(gauge, val)
					 return nil
				 }, gauge,
			)	

			if err != nil {
				return err
			}
		} else {
			switch v := rawVal.(type) {
			case string:
				fmt.Printf("Attribute: %s = %s\n", key, v)
			case []interface{}:
				strSlice := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						strSlice = append(strSlice, s)
					}
				}
				fmt.Printf("List attribute: %s = %v\n", key, strSlice)
			default:
				fmt.Printf("Ignoring non-metric field: %s (%T)\n", key, val)
			}
		}
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

