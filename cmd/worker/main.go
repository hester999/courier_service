package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	cfg "avito/internal/config"
	ordergate "avito/internal/gateway/order"
	eventhandler "avito/internal/handler/kafka"
	"avito/internal/infra/db"
	"avito/internal/infra/monitoring"
	cr "avito/internal/repo/courier"
	dr "avito/internal/repo/delivery"

	du "avito/internal/usecases/delivery"
	deadfactory "avito/internal/usecases/delivery/factory"
	"avito/internal/usecases/event"
	eventfactory "avito/internal/usecases/event/factory"
	"avito/pkg"

	"github.com/IBM/sarama"
)

func main() {
	log.Println("Starting worker...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	logger := pkg.NewLogger(0)
	cfg := cfg.NewConfig(logger)
	cfg.SetCFGFromEnvFile()
	cfg.LoadConfig()

	broker := cfg.KafkaHost
	topic := cfg.KafkaTopic
	groupID := "delivery_group"

	conn, err := db.InitDataBaseConnection(cfg.MaxAttempts, cfg.Delay)
	if err != nil {
		log.Printf("db connect error: %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	log.Println("kafka: topic:", broker, topic)
	factory := trmsqlx.NewDefaultFactory(conn)
	mgr := manager.Must(factory)
	getter := trmsqlx.DefaultCtxGetter
	txManager := db.NewTRMRunner(mgr)

	monitoring.Init()
	courStorage := cr.NewRepoCourier(conn, getter, logger)
	deliveryStorage := dr.NewDeliveryRepo(conn, getter, logger)

	deadlineFactory := deadfactory.NewDeadLineFabric()
	gateWay, _ := ordergate.NewGRPCOrderGateway(cfg.OrderServiceAddress, cfg.ConnRetryToOrderService, cfg.ConnectDelay, cfg.RetryDelay, monitoring.GateWayRetryTotal, logger)

	deliveryUsecases := du.NewDeliveryUsecase(txManager, deliveryStorage, courStorage, deadlineFactory, gateWay, logger)

	handlerFactory := eventfactory.NewHandlerFactory(deliveryUsecases, logger)
	eventUsecases := event.NewEvent(handlerFactory, gateWay, logger)

	eventHandler := eventhandler.NewKafkaHandler(eventUsecases, logger)

	kafkaClient, err := sarama.NewConsumerGroup([]string{broker}, groupID, config)
	if err != nil {
		log.Println("Error creating consumer group client:", err)
		return
	}

	go func() {
		for {
			err := kafkaClient.Consume(ctx, []string{topic}, eventHandler)
			if err != nil {
				log.Printf("consume error: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	log.Println("Event worker started")

	waitGracefulShutdown(cancel)

	log.Println("Event worker stopped gracefully.")
}

func waitGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var signal string
	// select {
	// case sig := <-sigChan:
	// 	signal = sig.String()
	// }
	sig := <-sigChan
	signal = sig.String()

	cancel()

	log.Printf("Shutdown signal (%s)", signal)
}
