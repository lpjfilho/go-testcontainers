package post_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"testcontainers/internal/infra/repository"
	"testcontainers/internal/post"
	"testcontainers/pkg/queue"
	"testcontainers/test/integration"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/goleak"
)

const (
	sqlPath = "../../test/testdata/sql"
)

var postExpected *post.Post
var aJson []byte

var postRepository post.PostRepositoryInterface
var postService post.PostServiceInterface

type PreviewSuite struct {
	suite.Suite
	Ctx               context.Context
	SqlConn           *sql.DB
	PostgresContainer *integration.Postgres
	RabbitMQ          *queue.RabbitMQ
	RabbitMQContainer *integration.RabbitMQ
}

func TestPreviewSuite(t *testing.T) {
	suite.Run(t, new(PreviewSuite))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func (s *PreviewSuite) SetupSuite() {
	var err error

	ctx, cancelFn := context.WithTimeout(context.TODO(), 3*time.Minute)
	defer cancelFn()

	s.PostgresContainer = integration.NewPostgres(s.T(), sqlPath)
	s.SqlConn = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(s.PostgresContainer.DSN(s.T()))))
	s.Require().NoError(s.SqlConn.PingContext(ctx))

	s.RabbitMQContainer = integration.NewRabbitMQ(s.T())

	port := s.RabbitMQContainer.Port(s.T())

	s.RabbitMQ, err = queue.NewRabbitMQ(queue.Config{
		User:       s.RabbitMQContainer.User,
		Pass:       s.RabbitMQContainer.Pass,
		Host:       s.RabbitMQContainer.Host(s.T()),
		Port:       strconv.Itoa(port),
		Exchange:   "post",
		Queue:      "post",
		RoutingKey: "",
	})

	s.Require().NoError(err)
}

func (s *PreviewSuite) TearDownSuite() {
	s.SqlConn.Close()
	s.PostgresContainer.Close(s.T())

	s.RabbitMQ.Close()
	s.RabbitMQContainer.Close(s.T())
}

func (s *PreviewSuite) SetupTest() {
	s.Ctx = context.TODO()
	var err error

	postExpected, err = post.NewPost("Test Containers", "Go test containers")
	s.Require().NoError(err)

	aJson, err = json.Marshal(postExpected)
	s.Require().NoError(err)

	postRepository = repository.NewPostRepository(s.SqlConn)
	postService = post.NewPostService(postRepository)
}

func (s *PreviewSuite) TestSaveAndFind() {
	err := s.RabbitMQ.Publish(s.Ctx, string(aJson))
	s.Require().NoError(err)

	proc := make(chan map[bool]amqp.Delivery, 10)
	go func() {
		for p := range proc {
			for k, v := range p {
				if k {
					v.Ack(false)
				} else {
					v.Reject(false)
				}
			}
		}
	}()

	msgs, err := s.RabbitMQ.Consume()
	s.Require().NoError(err)

	msg := <-msgs
	proc <- map[bool]amqp.Delivery{true: msg}
	close(proc)

	var p post.Post
	err = json.Unmarshal(msg.Body, &p)

	s.Require().NoError(err)
	s.Assert().Equal(*postExpected, p)

	err = postService.Create(s.Ctx, &p)
	s.Require().NoError(err)

	res, err := postService.Find(s.Ctx, p.ID)

	s.Require().NoError(err)
	s.Assert().Equal(*postExpected, *res)
}
