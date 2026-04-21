package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	pb "github.com/viacheslavprokosa/gocron/proto/task/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func sendTask(addr string, payload string) {
	con, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer con.Close()

	client := pb.NewTaskServiceClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	taskId := "cron-" + time.Now().Format("150405")

	res, err := client.RunTask(ctx, &pb.TaskRequest{
		TaskId:  taskId,
		Payload: payload,
	})
	if err != nil {
		log.Fatalf("Failed to run task: %v", err)
		return
	}
	log.Printf("Task %s : %v - addr: %s", res.Message, res.Success, addr)
}

func main() {
	addr := flag.String("addr", "", "Address of the gRPC microservice")
	schedule := flag.String("cron", "", "Cron schedule (e.g '0 * * * *' or '@every 1m')")
	payload := flag.String("payload", "", "Payload to send to the gRPC microservice")
	flag.Parse()

	if *addr == "" || *schedule == "" {
		fmt.Println("Usage: gocron --addr=<grpc-address> --cron=<cron-schedule>")
		os.Exit(2)
	}

	c := cron.New()

	_, err := c.AddFunc(*schedule, func() {
		log.Printf("Triggering job for %s", *addr)
		sendTask(*addr, *payload)
	})
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}
	c.Start()
	log.Printf("Cron scheduler started for %s with schedule %s and payload %s", *addr, *schedule, *payload)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down")
	c.Stop()
}
