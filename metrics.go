package main 

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type rawMetric map[string]interface{}

type metricSample struct {
	labels map[string]string
	value  float64
}

type MetricsStore struct {
	mu      sync.RWMutex
	metrics map[string][]metricSample
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		metrics: make(map[string][]metricSample),
	}
}

func (ms *MetricsStore) UpdateMetrics(samples []metricSample, metricName string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.metrics[metricName] = samples
}

func (ms *MetricsStore) GetMetrics() map[string][]metricSample {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	cpy := make(map[string][]metricSample, len(ms.metrics))
	for k, v := range ms.metrics {
		cpy[k] = v
	}
	return cpy
}

var store = NewMetricsStore()

func RecordMetrics(ctx context.Context, meter metric.Meter, data []byte) error {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}

	var rawMetrics []rawMetric
	switch v := parsed.(type) {
	case []interface{}:
		for i, e := range v {
			obj, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("element %d is not an object", i)
			}
			rawMetrics = append(rawMetrics, obj)
		}
	case map[string]interface{}:
		rawMetrics = append(rawMetrics, v)
	default:
		return errors.New("json must be object or array of objects")
	}

	batchByName := make(map[string][]metricSample)

	for i, rm := range rawMetrics {
		metricNameRaw, ok := rm["metricName"]
		if !ok {
			return fmt.Errorf("metricName missing in object %d", i)
		}
		metricName, ok := metricNameRaw.(string)
		if !ok || metricName == "" {
			return fmt.Errorf("metricName must be non-empty string in object %d", i)
		}

		valueRaw, ok := rm["value"]
		if !ok {
			return fmt.Errorf("value missing in object %d", i)
		}

		labels := make(map[string]string)
		for k, v := range rm {
			if k == "metricName" || k == "value" {
				continue
			}
			labels[k] = stringifyLabelValue(v)
		}

		var val float64
		switch vv := valueRaw.(type) {
		case float64:
			val = vv
		case int64:
			val = float64(vv)
		case int:
			val = float64(vv)
		case json.Number:
			fv, err := vv.Float64()
			if err != nil {
				iv, err2 := vv.Int64()
				if err2 != nil {
					return fmt.Errorf("invalid json.Number in object %d: %v", i, vv)
				}
				val = float64(iv)
			} else {
				val = fv
			}
		case string:
			labels["value"] = vv
			val = 1
		case []interface{}:
			var parts []string
			for _, x := range vv {
				parts = append(parts, stringifyLabelValue(x))
			}
			labels["value"] = strings.Join(parts, ",")
			val = 1
		default:
			return fmt.Errorf("unsupported value type %T in object %d", vv, i)
		}

		batchByName[metricName] = append(batchByName[metricName], metricSample{
			labels: labels,
			value:  val,
		})
	}

	for metricName, samples := range batchByName {
		store.UpdateMetrics(samples, metricName)

		gauge, err := meter.Float64ObservableGauge(metricName)
		if err != nil {
			return fmt.Errorf("failed to create gauge %s: %w", metricName, err)
		}

		_, err = meter.RegisterCallback(
			func(ctx context.Context, observer metric.Observer) error {
				metrics := store.GetMetrics()
				samplesForMetric, ok := metrics[metricName]
				if !ok {
					return nil
				}
				for _, sample := range samplesForMetric {
					attrs := labelsToAttributes(sample.labels)
					opts := []metric.ObserveOption{metric.WithAttributes(attrs...)}
					observer.ObserveFloat64(gauge, sample.value, opts...)
				}
				return nil
			},
			gauge,
		)
		if err != nil {
			return fmt.Errorf("failed to register callback for %s: %w", metricName, err)
		}
	}

	return nil
}

func stringifyLabelValue(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return vv
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64)
	case int:
		return strconv.Itoa(vv)
	case int64:
		return strconv.FormatInt(vv, 10)
	case json.Number:
		return vv.String()
	case []interface{}:
		var parts []string
		for _, x := range vv {
			parts = append(parts, stringifyLabelValue(x))
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", vv)
	}
}

func labelsToAttributes(labels map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels))
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		attrs = append(attrs, attribute.String(k, labels[k]))
	}
	return attrs
}

