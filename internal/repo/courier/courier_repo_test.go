package courier_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/repo/courier"
	"avito/pkg"
)

func newTestRepoCourier(t *testing.T) (*courier.RepoCourier, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := pkg.NewLogger(0)

	repo := courier.NewRepoCourier(sqlxDB, trmsqlx.DefaultCtxGetter, logger)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}

	return repo, mock, cleanup
}

func TestRepoCourier_GetCourierById_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	}).AddRow(
		int64(1),
		"Ivan",
		"+79000000000",
		"available",
		now,
		now,
		"car",
	)

	mock.
		ExpectQuery(regexp.QuoteMeta(
			"select id,name,phone,status,created_at,updated_at,transport_type from couriers where id = $1",
		)).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	got, err := repo.GetCourierById(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, int64(1), got.ID)
	assert.Equal(t, "Ivan", got.Name)
	assert.Equal(t, "+79000000000", got.Phone)
	assert.Equal(t, "available", got.Status)
	require.NotNil(t, got.TransportType)
	assert.Equal(t, "car", *got.TransportType)
}

func TestRepoCourier_GetCourierById_NotFound(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	mock.
		ExpectQuery(regexp.QuoteMeta(
			"select id,name,phone,status,created_at,updated_at,transport_type from couriers where id = $1",
		)).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetCourierById(ctx, 1)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "courierRepo not found", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_GetCourierById_InternalError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	dbErr := errors.New("some db error")

	mock.
		ExpectQuery(regexp.QuoteMeta(
			"select id,name,phone,status,created_at,updated_at,transport_type from couriers where id = $1",
		)).
		WithArgs(int64(1)).
		WillReturnError(dbErr)

	got, err := repo.GetCourierById(ctx, 1)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_GetCouriers_Success_NoLimitOffset(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	}).AddRow(
		int64(1),
		"Ivan",
		"+79000000000",
		"available",
		now,
		now,
		"car",
	)

	mock.
		ExpectQuery("SELECT .* FROM couriers").
		WithArgs().
		WillReturnRows(rows)

	result, err := repo.GetCouriers(ctx, nil, nil)
	require.NoError(t, err)
	require.Len(t, result, 1)

	c := result[0]
	assert.Equal(t, int64(1), c.ID)
	assert.Equal(t, "Ivan", c.Name)
	assert.Equal(t, "+79000000000", c.Phone)
	assert.Equal(t, "available", c.Status)
	require.NotNil(t, c.TransportType)
	assert.Equal(t, "car", *c.TransportType)
}

func TestRepoCourier_GetCouriers_Success_WithLimitOffset(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	limit := int64(10)
	offset := int64(5)

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	}).AddRow(
		int64(2),
		"Петя",
		"+79990001122",
		"busy",
		now,
		now,
		"scooter",
	)

	mock.
		ExpectQuery(regexp.QuoteMeta("SELECT id, name, phone, status, created_at, updated_at, transport_type FROM couriers LIMIT 10 OFFSET 5")).
		WithArgs().
		WillReturnRows(rows)

	result, err := repo.GetCouriers(ctx, &limit, &offset)
	require.NoError(t, err)
	require.Len(t, result, 1)

	c := result[0]
	assert.Equal(t, int64(2), c.ID)
	assert.Equal(t, "Петя", c.Name)
	assert.Equal(t, "+79990001122", c.Phone)
	assert.Equal(t, "busy", c.Status)
	require.NotNil(t, c.TransportType)
	assert.Equal(t, "scooter", *c.TransportType)
}

func TestRepoCourier_GetCouriers_DBError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	dbErr := errors.New("db is down")

	mock.
		ExpectQuery("SELECT .* FROM couriers").
		WithArgs().
		WillReturnError(dbErr)

	result, err := repo.GetCouriers(ctx, nil, nil)

	require.Nil(t, result)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_GetCouriers_NoCouriers(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	})

	mock.
		ExpectQuery("SELECT .* FROM couriers").
		WithArgs().
		WillReturnRows(rows)

	result, err := repo.GetCouriers(ctx, nil, nil)

	require.Nil(t, result)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "no couriers", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_CreateCourier_BeginError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	cour := &model.Courier{
		Name:   "Ivan",
		Phone:  "+79000000000",
		Status: "available",
	}

	mock.ExpectBegin().WillReturnError(errors.New("db down"))

	got, err := repo.CreateCourier(ctx, cour)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "internal server error", appErr.Message)
	assert.EqualError(t, appErr.ServiceError, "db down")
}

func TestRepoCourier_CreateCourier_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	transport := "car"
	cour := &model.Courier{
		Name:          "Ivan",
		Phone:         "+79000000000",
		Status:        "available",
		TransportType: &transport,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO couriers",
	)).
		WithArgs(
			cour.Name,
			cour.Phone,
			cour.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			*cour.TransportType,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
			}).AddRow(
				int64(1),
				cour.Name,
				cour.Phone,
				cour.Status,
				now,
				now,
				*cour.TransportType,
			),
		)

	mock.ExpectCommit()

	got, err := repo.CreateCourier(ctx, cour)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, int64(1), got.ID)
	assert.Equal(t, cour.Name, got.Name)
	assert.Equal(t, cour.Phone, got.Phone)
	assert.Equal(t, cour.Status, got.Status)
	require.NotNil(t, got.TransportType)
	assert.Equal(t, *cour.TransportType, *got.TransportType)
}

func TestRepoCourier_CreateCourier_DuplicatePhone(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	transport := "car"
	cour := &model.Courier{
		Name:          "Ivan",
		Phone:         "+79000000000",
		Status:        "available",
		TransportType: &transport,
	}

	mock.ExpectBegin()

	pgErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value violates unique constraint",
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO couriers",
	)).
		WithArgs(
			cour.Name,
			cour.Phone,
			cour.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			*cour.TransportType,
		).
		WillReturnError(pgErr)

	mock.ExpectRollback()

	got, err := repo.CreateCourier(ctx, cour)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "courier with this phone already exists", appErr.Message)
	assert.Equal(t, apperr.ErrAlreadyExists, appErr.ServiceError)
}

func TestRepoCourier_CreateCourier_CommitError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	transport := "car"
	cour := &model.Courier{
		Name:          "Ivan",
		Phone:         "+79000000000",
		Status:        "available",
		TransportType: &transport,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO couriers",
	)).
		WithArgs(
			cour.Name,
			cour.Phone,
			cour.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			*cour.TransportType,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
			}).AddRow(
				int64(1),
				cour.Name,
				cour.Phone,
				cour.Status,
				now,
				now,
				*cour.TransportType,
			),
		)

	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	got, err := repo.CreateCourier(ctx, cour)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "internal server error", appErr.Message)
	assert.EqualError(t, appErr.ServiceError, "commit failed")
}

func TestRepoCourier_UpdateCourier_BeginError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	cour := &model.Courier{ID: 1, Name: "Ivan"}

	beginErr := errors.New("db down")

	mock.ExpectBegin().WillReturnError(beginErr)

	res, err := repo.UpdateCourier(ctx, cour)
	require.Nil(t, res)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, beginErr, appErr.ServiceError)
}

func TestRepoCourier_UpdateCourier_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	transport := "car"
	cour := &model.Courier{
		ID:            1,
		Name:          "Ivan",
		Phone:         "+79000000000",
		Status:        "available",
		TransportType: &transport,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(
		"UPDATE couriers SET",
	)).
		WithArgs(
			cour.Name,
			cour.Phone,
			cour.Status,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			cour.ID,
		).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "name", "phone", "status", "created_at", "updated_at", "transport_type", "order_count",
			}).AddRow(
				int64(1),
				cour.Name,
				cour.Phone,
				cour.Status,
				now,
				now,
				transport,
				int64(5),
			),
		)

	mock.ExpectCommit()

	res, err := repo.UpdateCourier(ctx, cour)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, int64(1), res.ID)
	assert.Equal(t, "Ivan", res.Name)
	assert.Equal(t, "+79000000000", res.Phone)
	assert.Equal(t, "available", res.Status)
	require.NotNil(t, res.TransportType)
	assert.Equal(t, "car", *res.TransportType)
}

func TestRepoCourier_UpdateCourier_NotFound(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	c := &model.Courier{
		ID:     42,
		Status: "busy",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(
		"UPDATE couriers SET",
	)).
		WithArgs(
			"busy",
			sqlmock.AnyArg(),
			int64(42),
		).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	got, err := repo.UpdateCourier(ctx, c)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "courierRepo not found", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_UpdateCourier_InternalErrorOnQuery(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	c := &model.Courier{
		ID:     42,
		Status: "busy",
	}

	mock.ExpectBegin()

	dbErr := errors.New("db is down")

	mock.ExpectQuery(regexp.QuoteMeta(
		"UPDATE couriers SET",
	)).
		WithArgs(
			"busy",
			sqlmock.AnyArg(),
			int64(42),
		).
		WillReturnError(dbErr)

	mock.ExpectRollback()

	got, err := repo.UpdateCourier(ctx, c)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_GetCouriersByStatus_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	}).AddRow(
		int64(1),
		"Courier 1",
		"+79000000001",
		"available",
		now,
		now,
		"car",
	).AddRow(
		int64(2),
		"Courier 2",
		"+79000000002",
		"available",
		now,
		now,
		"scooter",
	)

	mock.ExpectQuery(regexp.QuoteMeta(
		"select id, name, phone, status, created_at, updated_at, transport_type from couriers where status = $1",
	)).
		WithArgs("available").
		WillReturnRows(rows)

	result, err := repo.GetCouriersByStatus(ctx, "available")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "Courier 1", result[0].Name)
	assert.Equal(t, "+79000000001", result[0].Phone)
	assert.Equal(t, "available", result[0].Status)
	require.NotNil(t, result[0].TransportType)
	assert.Equal(t, "car", *result[0].TransportType)

	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "Courier 2", result[1].Name)
	assert.Equal(t, "+79000000002", result[1].Phone)
	assert.Equal(t, "available", result[1].Status)
	require.NotNil(t, result[1].TransportType)
	assert.Equal(t, "scooter", *result[1].TransportType)
}

func TestRepoCourier_GetCouriersByStatus_DBError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	dbErr := errors.New("db is down")

	mock.ExpectQuery(regexp.QuoteMeta(
		"select id, name, phone, status, created_at, updated_at, transport_type from couriers where status = $1",
	)).
		WithArgs("busy").
		WillReturnError(dbErr)

	result, err := repo.GetCouriersByStatus(ctx, "busy")

	require.Nil(t, result)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_GetCouriersByStatus_NoCouriers(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "name", "phone", "status", "created_at", "updated_at", "transport_type",
	})

	mock.ExpectQuery(regexp.QuoteMeta(
		"select id, name, phone, status, created_at, updated_at, transport_type from couriers where status = $1",
	)).
		WithArgs("busy").
		WillReturnRows(rows)

	result, err := repo.GetCouriersByStatus(ctx, "busy")

	require.Nil(t, result)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "Couriers not found", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_GetCouriersWithMinimalOrders_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "transport_type",
	}).AddRow(
		int64(1), "car",
	).AddRow(
		int64(2), "scooter",
	)

	mock.ExpectQuery(`select\s+id,\s*transport_type\s+from\s+couriers\s+where\s+status\s*=\s*'available'`).
		WillReturnRows(rows)

	got, err := repo.GetCouriersWithMinimalOrders(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, int64(1), got[0].ID)
	require.NotNil(t, got[0].TransportType)
	assert.Equal(t, "car", *got[0].TransportType)

	assert.Equal(t, int64(2), got[1].ID)
	require.NotNil(t, got[1].TransportType)
	assert.Equal(t, "scooter", *got[1].TransportType)
}

func TestRepoCourier_GetCouriersWithMinimalOrders_DBError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	dbErr := errors.New("db is down")

	mock.ExpectQuery(`select\s+id,\s*transport_type\s+from\s+couriers\s+where\s+status\s*=\s*'available'`).
		WillReturnError(dbErr)

	got, err := repo.GetCouriersWithMinimalOrders(ctx)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_GetCouriersWithMinimalOrders_NoCouriers(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{
		"id", "transport_type",
	})

	mock.ExpectQuery(`select\s+id,\s*transport_type\s+from\s+couriers\s+where\s+status\s*=\s*'available'`).
		WillReturnRows(rows)

	got, err := repo.GetCouriersWithMinimalOrders(ctx)

	require.Nil(t, got)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "Couriers not found", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_IncrementTotalOrder_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	courierID := int64(1)

	mock.ExpectExec(regexp.QuoteMeta(
		"update couriers set order_count = order_count + 1 where id = $1",
	)).
		WithArgs(courierID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementTotalOrder(ctx, courierID)
	require.NoError(t, err)
}

func TestRepoCourier_IncrementTotalOrder_NotFound(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	courierID := int64(42)

	mock.ExpectExec(regexp.QuoteMeta(
		"update couriers set order_count = order_count + 1 where id = $1",
	)).
		WithArgs(courierID).
		WillReturnError(sql.ErrNoRows)

	err := repo.IncrementTotalOrder(ctx, courierID)

	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "Courier not found", appErr.Message)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestRepoCourier_IncrementTotalOrder_DBError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	courierID := int64(100)
	dbErr := errors.New("db is down")

	mock.ExpectExec(regexp.QuoteMeta(
		"update couriers set order_count = order_count + 1 where id = $1",
	)).
		WithArgs(courierID).
		WillReturnError(dbErr)

	err := repo.IncrementTotalOrder(ctx, courierID)

	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "internal server error", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}

func TestRepoCourier_UpdateCouriers_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	ids := []int64{1, 2}
	status := "available"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE couriers SET status")).
		WithArgs(status, sqlmock.AnyArg(), ids[0], ids[1]).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.UpdateCouriers(ctx, ids, status)
	require.NoError(t, err)
}

func TestRepoCourier_UpdateCouriers_DBError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoCourier(t)
	defer cleanup()

	ctx := context.Background()
	ids := []int64{10, 20}
	status := "busy"
	dbErr := errors.New("db is down")

	mock.ExpectExec(regexp.QuoteMeta("UPDATE couriers SET status")).
		WithArgs(status, sqlmock.AnyArg(), ids[0], ids[1]).
		WillReturnError(dbErr)

	err := repo.UpdateCouriers(ctx, ids, status)
	require.Error(t, err)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)

	assert.Equal(t, "failed to update couriers", appErr.Message)
	assert.Equal(t, dbErr, appErr.ServiceError)
}
