package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/stockflow/backend/internal/config"
	"github.com/stockflow/backend/internal/httpapi"
	"github.com/stockflow/backend/internal/migrate"
	"github.com/stockflow/backend/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "apply migrations and exit")
	flag.Parse()

	cfg := config.Load()
	ctx := context.Background()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer st.Close()

	if *migrateOnly {
		if err := migrate.Up(ctx, st.Pool()); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		log.Println("migrations applied")
		return
	}

	addr := ":" + cfg.Port
	log.Printf("stockflow backend listening on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.NewRouter(st)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
