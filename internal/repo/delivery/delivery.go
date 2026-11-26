package delivery

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/pkg"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	_ "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jmoiron/sqlx"
)

type DeliveryRepo struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	log     pkg.Logger
	builder sq.StatementBuilderType
}

type delivery struct {
	ID         int64     `db:"id"`
	CourierID  int64     `db:"courier_id"`
	OrderID    string    `db:"order_id"`
	AssignedAt time.Time `db:"assigned_at"`
	Deadline   time.Time `db:"deadline"`
}

type assignDelivery struct {
	CourierID int64     `db:"courier_id"`
	OrderID   string    `db:"order_id"`
	Deadline  time.Time `db:"deadline"`
}

type unassignDelivery struct {
	CourierID int64  `db:"courier_id"`
	OrderID   string `db:"order_id"`
	Status    string `db:"status"`
}

func NewDeliveryRepo(db *sqlx.DB, c *trmsqlx.CtxGetter, log pkg.Logger) *DeliveryRepo {
	return &DeliveryRepo{
		db:      db,
		getter:  c,
		log:     log,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TODO erros for user
func (d *DeliveryRepo) AssignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error) {
	var res assignDelivery
	query := `insert into delivery (courier_id, order_id, deadline) values ($1, $2, $3) returning order_id, courier_id,deadline`
	tx := d.getter.DefaultTrOrDB(ctx, d.db)
	err := tx.GetContext(ctx, &res, query, delivery.CourierID, delivery.OrderID, delivery.Deadline)
	if err != nil {
		d.log.Error("AssignCourier %v", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, &apperr.AppErr{
					Message:      "Courier already assigned on this order",
					ServiceError: apperr.ErrAlreadyExists,
				}
			}
		}
		return nil, &apperr.AppErr{
			Message:      "cannot assign delivery to courier",
			ServiceError: err,
		}
	}
	return &model.Delivery{
		OrderID:   res.OrderID,
		CourierID: res.CourierID,
		Deadline:  res.Deadline,
		Courier:   delivery.Courier,
	}, nil
}

// TODO erros for user
func (d *DeliveryRepo) UnassignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error) {
	query := `select courier_id, order_id, status from delivery inner join couriers c on c.id = delivery.courier_id where delivery.order_id = $1;`
	var res unassignDelivery
	tx := d.getter.DefaultTrOrDB(ctx, d.db)
	err := tx.GetContext(ctx, &res, query, delivery.OrderID)
	if err != nil {
		d.log.Error("unassign delivery failed %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &apperr.AppErr{
				Message:      "delivery not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return nil, &apperr.AppErr{
			Message:      "error unassigning courier",
			ServiceError: err,
		}
	}
	return &model.Delivery{
		CourierID: res.CourierID,
		OrderID:   delivery.OrderID,
		Status:    res.Status,
	}, nil
}

func (d *DeliveryRepo) DeleteDeliveryByOrderID(ctx context.Context, orderID string) error {
	query := `delete from delivery where order_id = $1`
	tx := d.getter.DefaultTrOrDB(ctx, d.db)
	_, err := tx.ExecContext(ctx, query, orderID)
	if err != nil {
		d.log.Error("DeleteDeliveryByOrderID %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &apperr.AppErr{
				Message:      "delivery not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return &apperr.AppErr{
			Message:      "error deleting delivery by orderID",
			ServiceError: err,
		}
	}
	return nil
}

func (d *DeliveryRepo) GetOverdueDeliveriesWithBusyCouriers(ctx context.Context) ([]int64, error) {
	var courierIds []int64

	query := `select  c.id as id from couriers c join delivery d ON d.courier_id = c.id where  c.status = 'busy' 
              group by c.id having max(d.deadline) < now();` //  Обновил запрос. Так как, теперь заказы висят в `delivery`, то выбираю по самому последнему дедлайну.
	// Оставил unique. Таким образом, одному заказу будет назначен точно один курьер.
	tx := d.getter.DefaultTrOrDB(ctx, d.db)
	err := tx.SelectContext(ctx, &courierIds, query)
	if err != nil {
		d.log.Error("GetOverdueDeliveriesWithBusyCouriers %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //todo: на сколько так правильно делать ?
		}
		return nil, err
	}

	return courierIds, nil
}

func (d *DeliveryRepo) DeleteDelivers(ctx context.Context, couriersIDs []int64) error {

	query, args, err := d.builder.Delete("delivery").Where(sq.Eq{"courier_id": couriersIDs}).ToSql()
	if err != nil {
		return &apperr.AppErr{
			Message:      "failed to build query",
			ServiceError: err,
		}
	}

	tx := d.getter.DefaultTrOrDB(ctx, d.db)
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		d.log.Error("DeleteDeliveryByOrderID %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &apperr.AppErr{
				Message:      "delivery not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return &apperr.AppErr{
			Message:      "error deleting delivery by orderID",
			ServiceError: err,
		}
	}
	return nil
}

func (d *DeliveryRepo) AssignCouriers(ctx context.Context, deliveries []*model.Delivery) ([]*model.Delivery, error) {
	if len(deliveries) == 0 {
		d.log.Error("AssignCouriers: deliveries is empty")
		return nil, &apperr.AppErr{
			Message:      "delivery is empty",
			ServiceError: apperr.ErrNotFound,
		}
	}

	builder := sq.StatementBuilder.
		Insert("delivery").
		Columns("courier_id", "order_id", "deadline").
		PlaceholderFormat(sq.Dollar)

	for _, v := range deliveries {
		builder = builder.Values(v.CourierID, v.OrderID, v.Deadline)
	}

	builder = builder.Suffix("on conflict (order_id) do nothing")

	query, args, err := builder.ToSql()
	if err != nil {
		d.log.Error("AssignCouriers: %v", err)
		return nil, err
	}

	tx := d.getter.DefaultTrOrDB(ctx, d.db)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		d.log.Error("AssignCouriers: %v", err)
		return nil, &apperr.AppErr{
			Message:      "failed to assign deliveries",
			ServiceError: err,
		}
	}

	return deliveries, nil
}

func (d *DeliveryRepo) GetDeliveryByOrderID(ctx context.Context, orderID string) (*model.Delivery, error) {
	var res delivery
	query := `select id,courier_id,order_id,assigned_at,deadline from delivery where order_id = $1`

	err := d.db.GetContext(ctx, &res, query, orderID)
	if err != nil {
		d.log.Error("GetDeliveryByOrderID %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &apperr.AppErr{
				Message:      "delivery not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return nil, &apperr.AppErr{
			Message:      "failed to get delivery by orderID",
			ServiceError: err,
		}
	}
	return &model.Delivery{
		ID:         res.ID,
		OrderID:    res.OrderID,
		CourierID:  res.CourierID,
		AssignedAt: res.AssignedAt,
		Deadline:   res.Deadline,
	}, nil
}
