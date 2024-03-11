package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arifmahmudrana/parking-lot/db"
)

const version = "1.0.0"

type application struct {
	infoLog, errorLog *log.Logger
	version           string
	dbRepo            *db.DB // TODO: use interface for testing
}

func (app *application) ConnectDB() {
	dbRepo, err := db.NewDB(os.Getenv("MYSQL_DSN"))
	if err != nil {
		app.errorLog.Fatal(err)
	}

	app.dbRepo = dbRepo
}

// To run the application compile or run `MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/parking_lot' go run cmd/*.go“
func main() {
	app := &application{
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog: log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		version:  version,
	}

	app.ConnectDB()
	defer app.dbRepo.Close()

	// The HTTP Server
	server := &http.Server{
		Addr:              ":8080",
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		// For not calling cancel(https://github.com/grpc/grpc-go/issues/1099)
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				app.errorLog.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			app.errorLog.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		app.errorLog.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
