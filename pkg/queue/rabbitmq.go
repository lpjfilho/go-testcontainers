package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	User       string
	Pass       string
	Host       string
	Port       string
	Exchange   string
	Queue      string
	RoutingKey string
}

type RabbitMQ struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	config Config
}

func NewRabbitMQ(config Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.User, config.Pass, config.Host, config.Port))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := createQueue(ch, config.Exchange, config.Queue, config.RoutingKey); err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn:   conn,
		ch:     ch,
		config: config,
	}, nil
}

func createQueue(ch *amqp.Channel, exchange, queue, routingKey string) error {
	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.config.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	deliveries := make(chan amqp.Delivery)
	go func() {
		for msg := range msgs {
			deliveries <- msg
		}
	}()

	return (<-chan amqp.Delivery)(deliveries), nil
}

func (r *RabbitMQ) Publish(ctx context.Context, message string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.ch.PublishWithContext(
		ctx,
		r.config.Exchange,
		r.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Body: []byte(message),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.conn.Close(); err != nil {
		return err
	}

	if err := r.ch.Close(); err != nil {
		return err
	}

	return nil
}
