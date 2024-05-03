package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/correctinho/correct-jwt-service-go/app"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	gin.DisableConsoleColor()
	r := gin.Default()

	app.Router(r)

	host := fmt.Sprintf(":%s", os.Getenv("GO_PORT"))

	server := &http.Server{
		Addr:         host,
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		println(host)
		if e := server.ListenAndServe(); e != nil {
			panic(e)
		}
	}()

	// allProcesses.Wait()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)

}
