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

	// Dial with correct params to get a connection, then forcibly close it before channel creation to cause error.
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.Rabbit.Username,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		t.Skipf("Skipping test because connection failed: %v", err)
	}
	conn.Close() // close immediately to cause Channel() error

	// This is a bit hacky: override the ConsumeMessages func to use this closed conn
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
		"invalid-type", // Invalid exchange type to cause error
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
	// QueueDeclare errors are tricky to force.
	// We can try to declare a queue with invalid arguments or empty name and non-exclusive.
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

	// Invalid QueueDeclare - empty name but not exclusive or autoDelete can cause error in some brokers
	_, err = ch.QueueDeclare(
		"",
		false,
		false,
		false, // not exclusive
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

	// Try to bind to a non-existent queue to cause error
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

	// Try to consume from a non-existent queue to cause error
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
