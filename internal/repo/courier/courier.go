package courier

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"avito/pkg"

	"avito/internal/apperr"
	"avito/internal/model"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type RepoCourier struct {
	db      *sqlx.DB
	logger  pkg.Logger
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewRepoCourier(db *sqlx.DB, c *trmsqlx.CtxGetter, logger pkg.Logger) *RepoCourier {
	return &RepoCourier{
		db:      db,
		logger:  logger,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

type courierRepo struct {
	ID            int64     `db:"id"`
	Name          string    `db:"name"`
	Phone         string    `db:"phone"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	TransportType string    `db:"transport_type"`
	TotalOrders   int64     `db:"order_count"`
}

func (c *RepoCourier) GetCourierById(ctx context.Context, id int64) (*model.Courier, error) {
	query := `select id,name,phone,status,created_at,updated_at,transport_type from couriers where id = $1`

	var courier courierRepo
	err := c.db.GetContext(ctx, &courier, query, id)
	if err != nil {
		c.logger.Error("Get Courier by id error %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &apperr.AppErr{
				Message:      "courierRepo not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}

	return &model.Courier{
		ID:            courier.ID,
		Name:          courier.Name,
		Phone:         courier.Phone,
		Status:        courier.Status,
		CreatedAt:     courier.CreatedAt,
		UpdatedAt:     courier.UpdatedAt,
		TransportType: &courier.TransportType,
	}, nil
}

func (c *RepoCourier) GetCouriers(ctx context.Context, limit, offset *int64) ([]*model.Courier, error) {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("id", "name", "phone", "status", "created_at", "updated_at", "transport_type").
		From("couriers")

	if limit != nil {
		query = query.Limit(uint64(*limit))
	}
	if offset != nil {
		query = query.Offset(uint64(*offset))
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, &apperr.AppErr{
			Message:      "failed to build query",
			ServiceError: err,
		}
	}
	var courier []courierRepo
	err = c.db.SelectContext(ctx, &courier, sqlStr, args...)
	if err != nil {
		c.logger.Error("Get Couriers %v:", err)
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}

	if len(courier) == 0 {
		return nil, &apperr.AppErr{
			Message:      "no couriers",
			ServiceError: apperr.ErrNotFound,
		}
	}

	res := make([]*model.Courier, len(courier))

	for i, cour := range courier {
		res[i] = &model.Courier{
			ID:            cour.ID,
			Name:          cour.Name,
			Phone:         cour.Phone,
			Status:        cour.Status,
			CreatedAt:     cour.CreatedAt,
			UpdatedAt:     cour.UpdatedAt,
			TransportType: &cour.TransportType,
		}
	}
	return res, nil
}

func (c *RepoCourier) CreateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
	tx, err := c.db.Beginx()
	if err != nil {
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	defer func() {
		if err != nil {
			c.logger.Error("Rollback transaction.Error: %v", err)
			_ = tx.Rollback()
		}
	}()
	var createdCourier courierRepo
	createdCourier, err = c.createCourier(ctx, tx, *courier)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		c.logger.Error("Commit transaction error %v", err)
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	return &model.Courier{
		ID:            createdCourier.ID,
		Name:          createdCourier.Name,
		Phone:         createdCourier.Phone,
		Status:        createdCourier.Status,
		CreatedAt:     createdCourier.CreatedAt,
		UpdatedAt:     createdCourier.UpdatedAt,
		TransportType: &createdCourier.TransportType,
	}, nil
}

func (c *RepoCourier) createCourier(ctx context.Context, tx *sqlx.Tx, courier model.Courier) (courierRepo, error) {
	var res courierRepo

	query := c.builder.Insert("couriers").
		Columns("name", "phone", "status", "created_at", "updated_at")

	values := []interface{}{courier.Name, courier.Phone, courier.Status, time.Now(), time.Now()}

	if courier.TransportType != nil {
		query = query.Columns("transport_type")
		values = append(values, *courier.TransportType)
	}

	query = query.Values(values...).Suffix("RETURNING id, name, phone, status, created_at, updated_at, transport_type")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return courierRepo{}, &apperr.AppErr{
			Message:      "failed to build query",
			ServiceError: err,
		}
	}

	err = tx.GetContext(ctx, &res, sqlStr, args...)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				c.logger.Error("createCourier duplicate key error: %v", pgErr)
				return courierRepo{}, &apperr.AppErr{
					Message:      "courier with this phone already exists",
					ServiceError: apperr.ErrAlreadyExists,
				}
			}
		}

		c.logger.Error("createCourier %v", err)
		return courierRepo{}, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}

	return res, nil
}

func (c *RepoCourier) UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error) {
	tx, err := c.db.Beginx()

	if err != nil {
		c.logger.Error("UpdateCourier %v", err)
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updatedCourier, err := c.updateCourier(ctx, tx, *courier)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		c.logger.Error("Commit transaction error: %v", err)
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	return &model.Courier{
		ID:            updatedCourier.ID,
		Name:          updatedCourier.Name,
		Phone:         updatedCourier.Phone,
		Status:        updatedCourier.Status,
		CreatedAt:     updatedCourier.CreatedAt,
		UpdatedAt:     updatedCourier.UpdatedAt,
		TransportType: &updatedCourier.TransportType,
	}, nil
}

func (c *RepoCourier) updateCourier(ctx context.Context, tx *sqlx.Tx, courier model.Courier) (courierRepo, error) {
	var res courierRepo

	query := c.builder.Update("couriers")

	if courier.Name != "" {
		query = query.Set("name", courier.Name)
	}
	if courier.Phone != "" {
		query = query.Set("phone", courier.Phone)
	}
	if courier.Status != "" {
		query = query.Set("status", courier.Status)
	}
	if courier.TransportType != nil {
		query = query.Set("transport_type", courier.TransportType)
	}

	query = query.Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": courier.ID}).
		Suffix("returning id, name, phone, status, created_at, updated_at, transport_type,order_count")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return courierRepo{}, &apperr.AppErr{
			Message:      "failed to build query",
			ServiceError: err,
		}
	}
	mgtx := c.getter.DefaultTrOrDB(ctx, tx)
	if err = mgtx.GetContext(ctx, &res, sqlStr, args...); err != nil {
		c.logger.Error("UpdateCourier query error %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return courierRepo{}, &apperr.AppErr{
				Message:      "courierRepo not found",
				ServiceError: apperr.ErrNotFound,
			}
		}
		return courierRepo{}, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}

	return res, nil
}

func (c *RepoCourier) GetCouriersByStatus(ctx context.Context, status string) ([]*model.Courier, error) {
	var repoCours []courierRepo

	query := `select id, name, phone, status, created_at, updated_at, transport_type from couriers where status = $1`
	tx := c.getter.DefaultTrOrDB(ctx, c.db)
	err := tx.SelectContext(ctx, &repoCours, query, status)

	if err != nil {
		c.logger.Error("GetCouriersByStatus %v", err)
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}

	if len(repoCours) == 0 {
		return nil, &apperr.AppErr{
			Message:      "Couriers not found",
			ServiceError: apperr.ErrNotFound,
		}
	}

	res := make([]*model.Courier, len(repoCours))

	for i, cour := range repoCours {
		res[i] = &model.Courier{
			ID:            cour.ID,
			Name:          cour.Name,
			Phone:         cour.Phone,
			Status:        cour.Status,
			CreatedAt:     cour.CreatedAt,
			UpdatedAt:     cour.UpdatedAt,
			TransportType: &cour.TransportType,
		}
	}
	return res, nil
}

func (c *RepoCourier) GetCouriersWithMinimalOrders(ctx context.Context) ([]*model.Courier, error) {
	var couriers []*courierRepo

	query := `select id, transport_type from couriers where status = 'available' and order_count = (
        select min(order_count)
        from couriers
        where status = 'available');`

	tx := c.getter.DefaultTrOrDB(ctx, c.db)

	err := tx.SelectContext(ctx, &couriers, query)
	if err != nil {
		c.logger.Error("GetCouriersWithMinimalOrders: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &apperr.AppErr{
				Message:      "internal server error",
				ServiceError: err,
			}
		}
		return nil, &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	if len(couriers) == 0 {
		return nil, &apperr.AppErr{
			Message:      "Couriers not found",
			ServiceError: apperr.ErrNotFound,
		}
	}

	res := make([]*model.Courier, len(couriers))

	for i, cour := range couriers {
		res[i] = &model.Courier{
			ID:            cour.ID,
			Name:          cour.Name,
			Phone:         cour.Phone,
			Status:        cour.Status,
			CreatedAt:     cour.CreatedAt,
			UpdatedAt:     cour.UpdatedAt,
			TransportType: &cour.TransportType,
		}
	}
	return res, nil
}

func (c *RepoCourier) IncrementTotalOrder(ctx context.Context, courierID int64) error {
	query := `update couriers set order_count = order_count + 1 where id = $1`
	tx := c.getter.DefaultTrOrDB(ctx, c.db)
	_, err := tx.ExecContext(ctx, query, courierID)
	if err != nil {
		c.logger.Error("AddOrderCount %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return &apperr.AppErr{
				Message:      "Courier not found",
				ServiceError: apperr.ErrNotFound,
			}
		}

		return &apperr.AppErr{
			Message:      "internal server error",
			ServiceError: err,
		}
	}
	return nil
}

func (c *RepoCourier) UpdateCouriers(ctx context.Context, courierIDs []int64, status string) error {
	query, args, err := c.builder.
		Update("couriers").
		Set("status", status).
		Set("updated_at", time.Now()).
		Set("order_count", sq.Expr("order_count + 1")).
		Where(sq.Eq{"id": courierIDs}).
		ToSql()

	if err != nil {
		return &apperr.AppErr{
			Message:      "failed to build query",
			ServiceError: err,
		}
	}

	tx := c.getter.DefaultTrOrDB(ctx, c.db)
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		c.logger.Error("UpdateCouriers: %v", err)
		return &apperr.AppErr{
			Message:      "failed to update couriers",
			ServiceError: err,
		}
	}

	return nil
}
