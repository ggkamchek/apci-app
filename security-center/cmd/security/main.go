package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/black/apci-app/security-center/internal/admin"
	"github.com/black/apci-app/security-center/internal/config"
	"github.com/black/apci-app/security-center/internal/db"
	"github.com/black/apci-app/security-center/internal/graph"
	"github.com/black/apci-app/security-center/internal/ingest"
	"github.com/black/apci-app/security-center/internal/migrate"
	"github.com/black/apci-app/security-center/internal/ui"
	securityv1 "github.com/black/apci-app/security-center/pkg/pb/apci/security/v1"
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

	graphRepo := graph.NewRepository(pool)
	ingestService := ingest.NewService(graphRepo)
	ingestGRPC := ingest.NewGRPCServer(ingestService)
	adminHandler := admin.NewHandler(ingestService, cfg.AdminToken)

	grpcServer := grpc.NewServer()
	securityv1.RegisterSecurityIngestServer(grpcServer, ingestGRPC)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	go func() {
		log.Printf("security grpc listening on %s", cfg.GRPCAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(pool))
	mux.HandleFunc("/admin/links", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/stats", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/stats/top-clones", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/stats/activity-series", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/devices", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/account/", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/diff", adminHandler.ServeHTTP)
	mux.HandleFunc("/admin/lookup", adminHandler.Lookup)

	adminFS, err := fs.Sub(ui.Assets, "web")
	if err != nil {
		log.Fatalf("admin ui fs: %v", err)
	}
	adminStatic := http.StripPrefix("/admin", http.FileServer(http.FS(adminFS)))
	mux.Handle("/admin/", adminStatic)
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
	})

	httpServer := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("security http listening on %s (admin ui: /admin/)", cfg.ServerAddr)
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
		resp := healthResponse{Status: "ok", Database: "ok"}

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
