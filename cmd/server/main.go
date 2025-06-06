package main

import (
	"MatchingEngine/internal/kafka"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/handler"
	"MatchingEngine/internal/repository"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/internal/service"
	"MatchingEngine/internal/util"
	"MatchingEngine/orderBook"
)

func main() {
	fmt.Println("-----------------------------------------")
	fmt.Println("        ORDER BOOK & MATCHING ENGINE     ")
	fmt.Println("-----------------------------------------")

	config, err := util.LoadConfig("./integration/compose")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to the database:", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kafkaProducer := kafka.NewProducer(config.KafkaBroker, config.KafkaTopic)
	book := orderBook.NewOrderBook()
	book.KafkaProducer = kafkaProducer
	repo := repository.NewPostgresOrderRepository(sqlc.New(conn))
	asyncWriter := repository.NewAsyncDBWriter(repo, 10)
	orderService := service.NewOrderService(asyncWriter)
	requestHandler := handler.NewOrderRequestHandler(book, orderService)

	consumerOpts := rmq.ConsumerOpts{
		RabbitMQURL: config.RmqHost,
		QueueName:   "order_requests",
		Prefetch:    10,
	}
	consumer := rmq.NewConsumer(consumerOpts, requestHandler)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatalf("Failed to start consumer: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1) // Correctly define the channel
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
}
