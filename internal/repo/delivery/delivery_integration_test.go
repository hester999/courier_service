//go:build integration

package delivery_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"avito/internal/model"
	"avito/internal/repo/delivery"
	"avito/pkg"
)

type DeliverySuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *delivery.DeliveryRepo
}

func (s *DeliverySuite) SetupSuite() {
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
	s.repo = delivery.NewDeliveryRepo(s.db, getter, logger)
}

func (s *DeliverySuite) SetupTest() {
	_, err := s.db.Exec(`
		TRUNCATE delivery, couriers RESTART IDENTITY CASCADE;
	`)
	require.NoError(s.T(), err)
}

func (s *DeliverySuite) insertCourier(name, phone, status string) int64 {
	var id int64
	require.NoError(s.T(), s.db.QueryRow(`
		INSERT INTO couriers(name, phone, status)
		VALUES ($1,$2,$3) RETURNING id
	`, name, phone, status).Scan(&id))
	return id
}

func (s *DeliverySuite) newOrderID() string {
	return uuid.NewString()
}

func (s *DeliverySuite) TestAssignCourier_Success() {
	c := s.insertCourier("Ivan", "123", "available")
	o := s.newOrderID()

	d := &model.Delivery{
		CourierID: c,
		OrderID:   o,
		Deadline:  time.Now().Add(time.Hour),
	}

	res, err := s.repo.AssignCourier(context.Background(), d)
	s.NoError(err)
	s.Equal(c, res.CourierID)
	s.Equal(o, res.OrderID)
}

func (s *DeliverySuite) TestAssignCourier_Duplicate() {
	c := s.insertCourier("Ivan", "123", "available")
	o := s.newOrderID()

	d := &model.Delivery{CourierID: c, OrderID: o, Deadline: time.Now()}

	_, err := s.repo.AssignCourier(context.Background(), d)
	s.NoError(err)

	_, err = s.repo.AssignCourier(context.Background(), d)
	s.Error(err) // ONLY delivery duplicate, NOT couriers.phone
}

func (s *DeliverySuite) TestUnassignCourier_Success() {
	c := s.insertCourier("Ivan", "123", "busy")
	o := s.newOrderID()

	_, err := s.db.Exec(`
		INSERT INTO delivery(courier_id,order_id,deadline)
		VALUES ($1,$2,now())
	`, c, o)
	require.NoError(s.T(), err)

	res, err := s.repo.UnassignCourier(context.Background(), &model.Delivery{OrderID: o})
	s.NoError(err)
	s.Equal(c, res.CourierID)
}

func (s *DeliverySuite) TestUnassignCourier_NotFound() {
	_, err := s.repo.UnassignCourier(context.Background(),
		&model.Delivery{OrderID: uuid.NewString()})
	s.Error(err)
}

func (s *DeliverySuite) TestDeleteDeliveryByOrderID() {
	c := s.insertCourier("Ivan", "123", "available")
	o := s.newOrderID()

	_, err := s.db.Exec(`
		INSERT INTO delivery(courier_id,order_id,deadline)
		VALUES ($1,$2,now())
	`, c, o)
	require.NoError(s.T(), err)

	s.NoError(s.repo.DeleteDeliveryByOrderID(context.Background(), o))

	var count int
	require.NoError(s.T(), s.db.QueryRow(`SELECT COUNT(*) FROM delivery`).Scan(&count))
	s.Equal(0, count)
}

func (s *DeliverySuite) TestGetOverdueDeliveriesWithBusyCouriers_Success() {
	busy := s.insertCourier("Busy", "111", "busy")
	free := s.insertCourier("Free", "222", "available")

	o1 := s.newOrderID()
	o2 := s.newOrderID()

	_, err := s.db.Exec(`
		INSERT INTO delivery(courier_id,order_id,deadline)
		VALUES ($1,$2,now()-interval '1h')
	`, busy, o1)
	require.NoError(s.T(), err)

	_, err = s.db.Exec(`
		INSERT INTO delivery(courier_id,order_id,deadline)
		VALUES ($1,$2,now()-interval '1h')
	`, free, o2)
	require.NoError(s.T(), err)

	ids, err := s.repo.GetOverdueDeliveriesWithBusyCouriers(context.Background())
	s.NoError(err)
	s.Equal([]int64{busy}, ids)
}

func TestDeliverySuite(t *testing.T) {
	suite.Run(t, new(DeliverySuite))
}
