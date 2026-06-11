package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	usersv1 "github.com/black/apci-app/server/pkg/pb/apci/users/v1"
	"github.com/black/apci-app/server/internal/config"
	"github.com/black/apci-app/server/internal/db"
	"github.com/black/apci-app/server/internal/migrate"
	"github.com/black/apci-app/server/internal/securitycenter"
	"github.com/black/apci-app/server/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, cfg.MigrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	securityPublisher, err := securitycenter.NewPublisher(cfg.SecurityCenterAddr)
	if err != nil {
		log.Fatalf("security center client: %v", err)
	}

	userRepo := users.NewRepository(pool)
	userService := users.NewService(userRepo, cfg.ChallengeTTL, cfg.SessionTTL)
	userGRPC := users.NewGRPCServer(userService, securityPublisher)

	grpcServer := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(grpcServer, userGRPC)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	go func() {
		log.Printf("grpc listening on %s", cfg.GRPCAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(pool))

	httpServer := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("http listening on %s", cfg.ServerAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			Status:   "ok",
			Database: "ok",
		}

		if err := db.Ping(r.Context(), pool); err != nil {
			resp.Status = "degraded"
			resp.Database = "error"
			writeJSON(w, http.StatusServiceUnavailable, resp)
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
