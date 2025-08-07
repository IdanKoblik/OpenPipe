package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func InitMeterProvider() (*metric.MeterProvider, error) {
	 exporter, err := prometheus.New()
	 if err != nil {
	 	return nil, err
	 }
 
	 provider := metric.NewMeterProvider(metric.WithReader(exporter))
	 return provider, nil
}

func main() {
	 foundDocker := false
    for _, arg := range os.Args[1:] { 
		 if arg == "--docker" {
            foundDocker = true
            break
        }
    }

	 configFile := "config.yml"
	 if foundDocker {
		configFile = "/home/container/config.yml" 
	 }

	 cfg, err := ReadConfig(configFile)
	 if err != nil {
	 	 log.Fatalf("Cannot parse config file: %v", err)
	 }
 
	 http.Handle("/metrics", promhttp.Handler())
	 log.Printf("Serving metrics at http://%s:%d/metrics\n", cfg.Web.Host, cfg.Web.Port)
	 port := fmt.Sprintf(":%d", cfg.Web.Port)
 
	 go func() {
	 	 if err := http.ListenAndServe(port, nil); err != nil {
	 	 	 log.Fatalf("Failed to start HTTP server: %v", err)
	 	 }
	 }()
 
	 meterProvider, err := InitMeterProvider()
	 if err != nil {
	 	 log.Fatalf("Failed to initialize meter provider: %v", err)
	 }

	 defer func() {
	 	 if err := meterProvider.Shutdown(context.Background()); err != nil {
	 	 	 log.Fatalf("Failed to shut down meter provider: %v", err)
	 	 }
	 }()
	 otel.SetMeterProvider(meterProvider)

	 msgs, err := ConsumeMessages(cfg)
	 if err != nil {
		 log.Fatalf("Failed to start consuming messages: %s", err)
	 }

	 meter := otel.Meter("server_metrics")
	 log.Println("Waiting for messages...")
	 for msg := range msgs {
	 	 log.Printf("Received: %s", msg.Body)
	 	 msg.Ack(false)
		 if err := RecordMetrics(context.Background(), meter, msg.Body); err != nil {
			fmt.Println("Error recording metrics:", err)
		 }
	 }
}

