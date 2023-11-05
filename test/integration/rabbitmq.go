package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type RabbitMQ struct {
	User     string
	Pass     string
	instance testcontainers.Container
}

func NewRabbitMQ(t *testing.T) *RabbitMQ {
	t.Helper()

	timeout := 3 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rabbitmqContainer, err := rabbitmq.RunContainer(
		ctx,
		testcontainers.WithImage("rabbitmq:3.7.25-management-alpine"),
		rabbitmq.WithAdminUsername("admin"),
		rabbitmq.WithAdminPassword("password"),
	)

	require.NoError(t, err)

	return &RabbitMQ{
		User:     rabbitmqContainer.AdminUsername,
		Pass:     rabbitmqContainer.AdminPassword,
		instance: rabbitmqContainer,
	}
}

func (r *RabbitMQ) Host(t *testing.T) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := r.instance.Host(ctx)
	require.NoError(t, err)
	return p
}

func (r *RabbitMQ) Port(t *testing.T) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := r.instance.MappedPort(ctx, "5672")
	require.NoError(t, err)
	return p.Int()
}

func (r *RabbitMQ) Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	require.NoError(t, r.instance.Terminate(ctx))
}
