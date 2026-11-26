package delivery_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/repo/delivery"
	"avito/pkg"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepoDelivery(t *testing.T) (*delivery.DeliveryRepo, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	logger := pkg.NewLogger(0)

	repo := delivery.NewDeliveryRepo(sqlxDB, trmsqlx.DefaultCtxGetter, logger)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}

	return repo, mock, cleanup
}

func TestAssignCourier_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	dl := &model.Delivery{
		OrderID:   "o1",
		CourierID: 10,
		Deadline:  time.Now().Add(2 * time.Hour),
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		"insert into delivery (courier_id, order_id, deadline) values ($1, $2, $3) returning order_id, courier_id,deadline",
	)).WithArgs(dl.CourierID, dl.OrderID, dl.Deadline).
		WillReturnRows(sqlmock.NewRows([]string{"order_id", "courier_id", "deadline"}).AddRow(dl.OrderID, dl.CourierID, dl.Deadline))

	got, err := repo.AssignCourier(ctx, dl)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, dl.OrderID, got.OrderID)
	assert.Equal(t, dl.CourierID, got.CourierID)
}

func TestAssignCourier_AlreadyExists(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	dl := &model.Delivery{OrderID: "o1", CourierID: 10, Deadline: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta(
		"insert into delivery (courier_id, order_id, deadline) values ($1, $2, $3) returning order_id, courier_id,deadline",
	)).WithArgs(dl.CourierID, dl.OrderID, dl.Deadline).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	got, err := repo.AssignCourier(ctx, dl)
	require.Error(t, err)
	require.Nil(t, got)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.ErrAlreadyExists, appErr.ServiceError)
}

func TestUnassignCourier_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	dl := &model.Delivery{OrderID: "o42"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"select courier_id, order_id, status from delivery inner join couriers c on c.id = delivery.courier_id where delivery.order_id = $1;",
	)).WithArgs(dl.OrderID).
		WillReturnRows(sqlmock.NewRows([]string{"courier_id", "order_id", "status"}).AddRow(int64(5), dl.OrderID, "busy"))

	got, err := repo.UnassignCourier(ctx, dl)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, int64(5), got.CourierID)
	assert.Equal(t, dl.OrderID, got.OrderID)
	assert.Equal(t, "busy", got.Status)
}

func TestUnassignCourier_NotFound(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	dl := &model.Delivery{OrderID: "o_not"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"select courier_id, order_id, status from delivery inner join couriers c on c.id = delivery.courier_id where delivery.order_id = $1;",
	)).WithArgs(dl.OrderID).WillReturnError(sql.ErrNoRows)

	got, err := repo.UnassignCourier(ctx, dl)
	require.Error(t, err)
	require.Nil(t, got)

	var appErr *apperr.AppErr
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.ErrNotFound, appErr.ServiceError)
}

func TestDeleteDeliveryByOrderID_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	orderID := "o_del"

	mock.ExpectExec(regexp.QuoteMeta("delete from delivery where order_id = $1")).WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteDeliveryByOrderID(ctx, orderID)
	require.NoError(t, err)
}

func TestDeleteDeliveryByOrderID_Error(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	orderID := "o_del"

	mock.ExpectExec(regexp.QuoteMeta("delete from delivery where order_id = $1")).WithArgs(orderID).
		WillReturnError(errors.New("db error"))

	err := repo.DeleteDeliveryByOrderID(ctx, orderID)
	require.Error(t, err)
}

func TestGetOverdueDeliveriesWithBusyCouriers_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectQuery(
		`(?i)select.+from\s+couriers\s+c\s+join\s+delivery\s+d.+having\s+max`,
	).WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2)),
	)

	ids, err := repo.GetOverdueDeliveriesWithBusyCouriers(ctx)
	require.NoError(t, err)
	require.Len(t, ids, 2)
	assert.Equal(t, int64(1), ids[0])
	assert.Equal(t, int64(2), ids[1])
}

func TestGetOverdueDeliveriesWithBusyCouriers_NoRows(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()

	mock.ExpectQuery(
		`(?i)select.+from\s+couriers\s+c\s+join\s+delivery\s+d.+having\s+max`,
	).WillReturnError(sql.ErrNoRows)

	ids, err := repo.GetOverdueDeliveriesWithBusyCouriers(ctx)
	require.NoError(t, err)
	require.Nil(t, ids)
}

func TestDeleteDelivers_Success(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	idsToDelete := []int64{1, 2}

	mock.ExpectExec(regexp.QuoteMeta("delivery")).WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.DeleteDelivers(ctx, idsToDelete)
	require.NoError(t, err)
}

func TestDeleteDelivers_BuilderError(t *testing.T) {
	t.Parallel()

	repo, mock, cleanup := newTestRepoDelivery(t)
	defer cleanup()

	ctx := context.Background()
	idsToDelete := []int64{1}

	mock.ExpectExec(regexp.QuoteMeta("delivery")).WillReturnError(errors.New("exec error"))

	err := repo.DeleteDelivers(ctx, idsToDelete)
	require.Error(t, err)
}
