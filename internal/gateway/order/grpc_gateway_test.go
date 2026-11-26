package order_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"avito/internal/gateway/order"
	"avito/internal/gateway/order/pb"
	"avito/pkg"
)

type ordersSrv struct {
	pb.UnimplementedOrdersServiceServer

	lastFrom *timestamppb.Timestamp

	orders []*pb.Order
	byID   map[string]*pb.Order
}

func (s *ordersSrv) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	s.lastFrom = req.GetFrom()
	return &pb.GetOrdersResponse{Orders: s.orders}, nil
}

func (s *ordersSrv) GetOrderById(ctx context.Context, req *pb.GetOrderByIdRequest) (*pb.GetOrderByIdResponse, error) {
	return &pb.GetOrderByIdResponse{Order: s.byID[req.GetId()]}, nil
}

func startGRPCServer(t *testing.T, srv pb.OrdersServiceServer) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	g := grpc.NewServer()
	pb.RegisterOrdersServiceServer(g, srv)

	go func() {
		_ = g.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		g.Stop()
		_ = lis.Close()
	}
}

var retryCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "gateway_retries_total_test",
	Help: "test counter",
})

func TestGRPCOrderGateway_GetNewOrders_SendsFromAndMapsResponse(t *testing.T) {
	created1 := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	created2 := time.Date(2025, 12, 28, 11, 0, 0, 0, time.UTC)
	logger := pkg.NewLogger(0)
	s := &ordersSrv{
		orders: []*pb.Order{
			{Id: "o1", CreatedAt: timestamppb.New(created1)},
			{Id: "o2", CreatedAt: timestamppb.New(created2)},
		},
		byID: map[string]*pb.Order{},
	}

	addr, stop := startGRPCServer(t, s)
	defer stop()

	gw, err := order.NewGRPCOrderGateway(addr, 0, 0, 10, retryCounter, logger)
	if err != nil {
		t.Fatalf("NewGRPCOrderGateway: %v", err)
	}
	defer func() { _ = gw.Close() }()

	from := time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC)

	got, err := gw.GetNewOrders(context.Background(), from)
	if err != nil {
		t.Fatalf("GetNewOrders: %v", err)
	}

	if s.lastFrom == nil {
		t.Fatalf("expected server to receive From, got nil")
	}
	if !s.lastFrom.AsTime().Equal(from) {
		t.Fatalf("expected From=%v, got %v", from, s.lastFrom.AsTime())
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(got))
	}
	if got[0].ID != "o1" || !got[0].CreatedAt.Equal(created1) {
		t.Fatalf("unexpected order[0]: %+v", got[0])
	}
	if got[1].ID != "o2" || !got[1].CreatedAt.Equal(created2) {
		t.Fatalf("unexpected order[1]: %+v", got[1])
	}
}

func TestGRPCOrderGateway_GetOrderByID_MapsResponse(t *testing.T) {
	created := time.Date(2025, 12, 28, 9, 30, 0, 0, time.UTC)

	s := &ordersSrv{
		orders: []*pb.Order{},
		byID: map[string]*pb.Order{
			"42": {Id: "42", CreatedAt: timestamppb.New(created), Status: "CONFIRMED"},
		},
	}

	addr, stop := startGRPCServer(t, s)
	defer stop()
	logger := pkg.NewLogger(0)
	gw, err := order.NewGRPCOrderGateway(addr, 0, 0, 10, retryCounter, logger)
	if err != nil {
		t.Fatalf("NewGRPCOrderGateway: %v", err)
	}
	defer func() { _ = gw.Close() }()

	o, err := gw.GetOrderByID(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetOrderByID: %v", err)
	}
	if o == nil {
		t.Fatalf("expected order, got nil")
	}
	if o.ID != "42" || !o.CreatedAt.Equal(created) || o.Status != "CONFIRMED" {
		t.Fatalf("unexpected mapped order: %+v", o)
	}
}

func TestGRPCOrderGateway_GetOrders_ReturnsPointers(t *testing.T) {
	created1 := time.Date(2025, 12, 28, 10, 0, 0, 0, time.UTC)
	created2 := time.Date(2025, 12, 28, 11, 0, 0, 0, time.UTC)

	s := &ordersSrv{
		orders: []*pb.Order{
			{Id: "a", CreatedAt: timestamppb.New(created1)},
			{Id: "b", CreatedAt: timestamppb.New(created2)},
		},
		byID: map[string]*pb.Order{},
	}

	addr, stop := startGRPCServer(t, s)
	defer stop()
	logger := pkg.NewLogger(0)
	gw, err := order.NewGRPCOrderGateway(addr, 0, 0, 10, retryCounter, logger)
	if err != nil {
		t.Fatalf("NewGRPCOrderGateway: %v", err)
	}
	defer func() { _ = gw.Close() }()

	orders, err := gw.GetOrders(context.Background())
	if err != nil {
		t.Fatalf("GetOrders: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if orders[0] == nil || orders[1] == nil {
		t.Fatalf("expected non-nil pointers")
	}
	if orders[0].ID != "a" || !orders[0].CreatedAt.Equal(created1) {
		t.Fatalf("unexpected orders[0]: %+v", orders[0])
	}
	if orders[1].ID != "b" || !orders[1].CreatedAt.Equal(created2) {
		t.Fatalf("unexpected orders[1]: %+v", orders[1])
	}
}

func TestGRPCOrderGateway_Close_NoError(t *testing.T) {
	s := &ordersSrv{orders: []*pb.Order{}, byID: map[string]*pb.Order{}}

	addr, stop := startGRPCServer(t, s)
	defer stop()
	logger := pkg.NewLogger(0)
	gw, err := order.NewGRPCOrderGateway(addr, 0, 0, 10, retryCounter, logger)
	if err != nil {
		t.Fatalf("NewGRPCOrderGateway: %v", err)
	}

	if err := gw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
