package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
)

func HealthHandler(w http.ResponseWriter, r *http.Request, cfg *Config) {
	if !CheckMetricsEndpoint(fmt.Sprintf("http://%s/metrics", CreateAddr(cfg))) {
		http.Error(w, "metrics endpoint not healthy", http.StatusServiceUnavailable)
		return
	}

	if !CheckRabbitConnection(CreateConnStr(cfg)) {
		http.Error(w, "rabbitmq connection failed", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}


func CheckMetricsEndpoint(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error checking /metrics: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("/metrics returned status: %v", resp.Status)
		return false
	}
	return true
}

func CheckRabbitConnection(url string) bool {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("RabbitMQ connection failed: %v", err)
		return false
	}
	defer conn.Close()
	return true
}

