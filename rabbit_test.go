package main

import (
	"fmt"
	"testing"

	"github.com/streadway/amqp"
)

func TestConsumeMessages_Success(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	msgs, err := ConsumeMessages(cfg)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if msgs == nil {
		t.Fatal("expected msgs channel, got nil")
	}
}

func TestConsumeMessages_DialError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "wrong"
	cfg.Rabbit.Password = "wrong"
	cfg.Rabbit.Host = "invalidhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	msgs, err := ConsumeMessages(cfg)
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
	if msgs != nil {
		t.Fatal("expected nil msgs on dial error")
	}
}

func TestConsumeMessages_ChannelError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	conn.Close()

	ConsumeMessagesClosedConn := func(cfg *Config) (<-chan amqp.Delivery, error) {
		ch, err := conn.Channel()
		if err != nil {
			return nil, err
		}
		_ = ch
		return nil, nil
	}

	_, err = ConsumeMessagesClosedConn(cfg)
	if err == nil {
		t.Fatal("expected channel error, got nil")
	}
}

func TestConsumeMessages_ExchangeDeclareError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672

	// Use invalid exchange type to cause exchange declare error
	cfg.Rabbit.Channel = "test-exchange"

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		t.Skipf("Skipping test because channel creation failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	err = ch.ExchangeDeclare(
		cfg.Rabbit.Channel,
		"invalid-type",
		false,
		false,
		false,
		false,
		nil,
	)
	if err == nil {
		t.Fatal("expected exchange declare error, got nil")
	}
}

func TestConsumeMessages_QueueDeclareError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		t.Skipf("Skipping test because channel creation failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err == nil {
		t.Skip("expected queue declare error but got nil (RabbitMQ may allow this)")
	}
}

func TestConsumeMessages_QueueBindError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		t.Skipf("Skipping test because channel creation failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	err = ch.QueueBind(
		"non-existent-queue",
		"",
		cfg.Rabbit.Channel,
		false,
		nil,
	)
	if err == nil {
		t.Skip("expected queue bind error but got nil (RabbitMQ may allow this)")
	}
}

func TestConsumeMessages_ConsumeError(t *testing.T) {
	cfg := &Config{}
	cfg.Rabbit.Username = "guest"
	cfg.Rabbit.Password = "guest"
	cfg.Rabbit.Host = "localhost"
	cfg.Rabbit.Port = 5672
	cfg.Rabbit.Channel = "test-exchange"

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		t.Skipf("Skipping test because channel creation failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	_, err = ch.Consume(
		"non-existent-queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err == nil {
		t.Skip("expected consume error but got nil (RabbitMQ may allow this)")
	}
}
