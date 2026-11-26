package order

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"avito/internal/gateway/order/pb"
	"avito/internal/model"
	"avito/pkg"
)

type GRPCOrderGateway struct {
	client pb.OrdersServiceClient
	conn   *grpc.ClientConn

	log pkg.Logger
}

func NewGRPCOrderGateway(addr string, connRetry int, delay int, retryDelay int, retryMetric prometheus.Counter, logger pkg.Logger) (*GRPCOrderGateway, error) {

	var conn *grpc.ClientConn
	var err error

	for attempt := 0; attempt <= connRetry; attempt++ {

		conn, err = grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithChainUnaryInterceptor(
				retry.UnaryClientInterceptor(
					retry.WithMax(5),
					retry.WithCodes(codes.ResourceExhausted, codes.Unavailable, codes.DeadlineExceeded),
					retry.WithBackoff(retry.BackoffExponential(time.Duration(retryDelay)*time.Millisecond)),
					retry.WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
						retryMetric.Inc()
						logger.Warn("[RETRY] attempt=%d err=%v", attempt, err)
					}),
				),
			),
		)
		if err == nil {
			break
		}

		log.Printf("failed to connect to grpc (%s), attempt %d/%d: %v",
			addr, attempt+1, connRetry+1, err)

		time.Sleep(time.Duration(delay) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC after %d retries: %w", connRetry, err)
	}

	client := pb.NewOrdersServiceClient(conn)

	return &GRPCOrderGateway{
		client: client,
		conn:   conn,
		log:    logger,
	}, nil
}

func (g *GRPCOrderGateway) Close() error {
	return g.conn.Close()
}

func (g *GRPCOrderGateway) GetNewOrders(ctx context.Context, from time.Time) ([]model.Order, error) {
	protoTime := timestamppb.New(from)

	res, err := g.client.GetOrders(ctx, &pb.GetOrdersRequest{
		From: protoTime,
	})
	if err != nil {
		return nil, err
	}
	var orders []model.Order
	for _, order := range res.Orders {
		orders = append(orders, model.Order{
			ID:        order.Id,
			CreatedAt: order.CreatedAt.AsTime(),
		})
	}
	return orders, nil
}

func (g *GRPCOrderGateway) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	res, err := g.client.GetOrderById(ctx, &pb.GetOrderByIdRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:        res.Order.Id,
		CreatedAt: res.Order.CreatedAt.AsTime(),
		Status:    res.Order.Status,
	}, nil
}

func (g *GRPCOrderGateway) GetOrders(ctx context.Context) ([]*model.Order, error) {
	res, err := g.client.GetOrders(ctx, &pb.GetOrdersRequest{})
	if err != nil {
		return nil, err
	}
	var orders []*model.Order
	for _, order := range res.Orders {
		orders = append(orders, &model.Order{
			ID:        order.Id,
			CreatedAt: order.CreatedAt.AsTime(),
		})
	}
	return orders, nil
}
