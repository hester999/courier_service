//go:build integration

package courier_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/repo/courier"
	"avito/pkg"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CourierSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *courier.RepoCourier
}

func (s *CourierSuite) SetupSuite() {
	host := os.Getenv("PG_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "9999"
	}
	user := os.Getenv("PG_USER")
	if user == "" {
		user = "test"
	}
	pass := os.Getenv("PG_PASSWORD")
	if pass == "" {
		pass = "test"
	}
	db := os.Getenv("PG_DB")
	if db == "" {
		db = "test"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)

	var err error
	s.db, err = sqlx.Connect("pgx", dsn)
	require.NoError(s.T(), err)

	logger := pkg.NewLogger(0)
	getter := trmsqlx.DefaultCtxGetter

	s.repo = courier.NewRepoCourier(s.db, getter, logger)
}

func (s *CourierSuite) SetupTest() {
	_, err := s.db.Exec(`
       TRUNCATE delivery, couriers RESTART IDENTITY CASCADE;
       SELECT setval('couriers_id_seq', 1, false);
   `)
	require.NoError(s.T(), err)
}

func (s *CourierSuite) insertTestCouriers() {
	_, err := s.db.Exec(`
       INSERT INTO couriers (name, phone, status)
       VALUES
           ('t1','123','available'),
           ('t2','1234','available'),
           ('t3','12345','available');
   `)
	require.NoError(s.T(), err)
}

func (s *CourierSuite) TestCreateCourier() {
	c, err := s.repo.CreateCourier(context.Background(),
		&model.Courier{Name: "Test", Phone: "1000", Status: "available"},
	)
	s.NoError(err)
	s.NotNil(c)

	_, err = s.repo.CreateCourier(context.Background(),
		&model.Courier{Name: "Dup", Phone: "1000", Status: "available"},
	)
	s.Error(err)

	var appErr *apperr.AppErr
	s.True(errors.As(err, &appErr))
	s.Equal(apperr.ErrAlreadyExists, appErr.ServiceError)
}

func (s *CourierSuite) TestGetCouriers() {
	s.insertTestCouriers()

	// limit=1 → t1
	res, err := s.repo.GetCouriers(context.Background(), ptr(int64(1)), ptr(int64(0)))
	s.NoError(err)
	s.Len(res, 1)
	s.Equal("t1", res[0].Name)

	// offset=1 → t2
	res, err = s.repo.GetCouriers(context.Background(), ptr(int64(1)), ptr(int64(1)))
	s.NoError(err)
	s.Len(res, 1)
	s.Equal("t2", res[0].Name)

	// limit=10 → all
	res, err = s.repo.GetCouriers(context.Background(), ptr(int64(10)), ptr(int64(0)))
	s.NoError(err)
	s.Len(res, 3)

	// limit=0 → error
	res, err = s.repo.GetCouriers(context.Background(), ptr(int64(0)), ptr(int64(0)))
	s.Error(err)
	s.Len(res, 0)
}

func (s *CourierSuite) TestGetCourierById() {
	s.insertTestCouriers()

	c, err := s.repo.GetCourierById(context.Background(), 1)
	s.NoError(err)
	s.Equal("t1", c.Name)

	_, err = s.repo.GetCourierById(context.Background(), 999)
	s.Error(err)

	var appErr *apperr.AppErr
	s.True(errors.As(err, &appErr))
	s.Equal(apperr.ErrNotFound, appErr.ServiceError)
}

func (s *CourierSuite) TestUpdateCourier() {
	_, err := s.repo.CreateCourier(context.Background(),
		&model.Courier{Name: "old", Phone: "111", Status: "available"},
	)
	s.NoError(err)

	u, err := s.repo.UpdateCourier(context.Background(),
		&model.Courier{ID: 1, Name: "new", Phone: "222", Status: "busy"},
	)
	s.NoError(err)
	s.Equal("new", u.Name)
	s.Equal("busy", u.Status)

	_, err = s.repo.UpdateCourier(context.Background(),
		&model.Courier{ID: 999},
	)
	s.Error(err)
}

func (s *CourierSuite) TestIncrementTotalOrder() {
	s.repo.CreateCourier(context.Background(),
		&model.Courier{Name: "c1", Phone: "1", Status: "available"},
	)

	s.NoError(s.repo.IncrementTotalOrder(context.Background(), 1))
	s.NoError(s.repo.IncrementTotalOrder(context.Background(), 1))

	var count int
	s.NoError(s.db.QueryRow(`SELECT order_count FROM couriers WHERE id=1`).Scan(&count))
	s.Equal(2, count)
}

func (s *CourierSuite) TestGetCouriersWithMinimalOrders() {
	s.insertTestCouriers()

	s.db.Exec(`UPDATE couriers SET order_count = 1 WHERE id = 3`)

	res, err := s.repo.GetCouriersWithMinimalOrders(context.Background())
	s.NoError(err)

	s.ElementsMatch([]int64{1, 2}, []int64{res[0].ID, res[1].ID})
}

func (s *CourierSuite) TestUpdateCouriers() {
	s.insertTestCouriers()

	err := s.repo.UpdateCouriers(context.Background(), []int64{1, 2}, "busy")
	s.NoError(err)

	rows, _ := s.db.Query(`SELECT status FROM couriers ORDER BY id`)
	var st []string
	for rows.Next() {
		var x string
		rows.Scan(&x)
		st = append(st, x)
	}
	s.Equal([]string{"busy", "busy", "available"}, st)
}

func TestCourierSuite(t *testing.T) {
	suite.Run(t, new(CourierSuite))
}

func ptr[T any](v T) *T { return &v }
