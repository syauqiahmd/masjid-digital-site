// Command http is the public site's API: the Daftar form (sent to the team's Telegram) and the
// "server bersama" numbers (read from IDCloudHost). Settings come from the environment, see
// .env.example.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // WIB without relying on the image's zoneinfo

	"github.com/syauqiahmd/masjid-digital-site/server/internal/config"
	deliveryhttp "github.com/syauqiahmd/masjid-digital-site/server/internal/delivery/http"
)

const gracefulTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	server := deliveryhttp.NewServer(cfg.HTTPServer, newController(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()

	select {
	case err := <-errs:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()
		server.Stop(shutdown)
	}
}
