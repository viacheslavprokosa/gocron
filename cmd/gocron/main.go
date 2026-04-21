package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
)

func main(){
	addr:=flag.String("addr","","Address of the gRPC microservice")
	schedule:=flag.String("cron","","Cron schedule (e.g '0 * * * *' or '@every 1m')")
	flag.Parse()

	if *addr=="" || *schedule==""{
		fmt.Println("Usage: gocron --addr=<grpc-address> --cron=<cron-schedule>")
		os.Exit(2)
	}

	c:=cron.New()

	_, err:=c.AddFunc(*schedule,func() {
		log.Printf("Triggering job for %s",*addr)
	})
	if err!=nil{
		log.Fatalf("Failed to add cron job: %v",err)
	}
	c.Start()
	log.Printf("Cron scheduler started for %s with schedule %s",*addr,*schedule)
	sigChan:=make(chan os.Signal,1)
	signal.Notify(sigChan,syscall.SIGINT,syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down")
	c.Stop()
}