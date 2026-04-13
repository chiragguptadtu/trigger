package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"trigger/db"
	"trigger/internal/auth"
	"trigger/internal/command"
	"trigger/internal/crypto"
	"trigger/internal/execution"
	"trigger/internal/handler"
	"trigger/internal/store"
)

func main() {
	_ = godotenv.Load()

	dbURL := mustEnv("DATABASE_URL")
	jwtSecret := mustEnv("JWT_SECRET")
	encKeyHex := mustEnv("ENCRYPTION_KEY")
	commandsDir := envOr("COMMANDS_DIR", "./commands")
	port := envOr("PORT", "8080")

	encKey, err := crypto.KeyFromHex(encKeyHex)
	if err != nil {
		log.Fatalf("invalid ENCRYPTION_KEY: %v", err)
	}

	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := db.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	q := store.New(pool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := seedAdminIfNeeded(ctx, q); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	if err := resetAdminIfRequested(ctx, q); err != nil {
		log.Fatalf("reset admin: %v", err)
	}

	// Auto-discover commands: scan immediately then every 30s.
	go command.ScanLoop(ctx, commandsDir, q, 30*time.Second)

	// River job queue
	worker := execution.NewWorker(q, encKey)
	riverClient, err := execution.NewRiverClient(pool, worker)
	if err != nil {
		log.Fatalf("river client: %v", err)
	}
	if err := riverClient.Start(ctx); err != nil {
		log.Fatalf("river start: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		_ = riverClient.Stop(stopCtx)
	}()

	enqueuer := execution.NewRiverEnqueuer(riverClient)
	srv := handler.NewServer(q, handler.Config{
		JWTSecret:     jwtSecret,
		TokenTTL:      int((24 * time.Hour).Seconds()),
		EncryptionKey: encKey,
	}, enqueuer)

	httpSrv := &http.Server{Addr: ":" + port, Handler: srv}

	go func() {
		log.Printf("listening on :%s", port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel() // stop scan loop and river client
	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

// seedAdminIfNeeded creates the first admin user when the DB is empty.
// Runs once on first deploy; no-op thereafter.
func seedAdminIfNeeded(ctx context.Context, q *store.Queries) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	users, err := q.ListUsers(ctx)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	_, err = q.CreateUser(ctx, store.CreateUserParams{
		Email:        email,
		Name:         "Admin",
		PasswordHash: hash,
		IsAdmin:      true,
	})
	if err != nil {
		return err
	}
	log.Printf("bootstrap admin created: %s", email)
	return nil
}

// resetAdminIfRequested resets an existing user's password on startup.
// Set RESET_ADMIN_EMAIL + RESET_ADMIN_PASSWORD to recover from admin lockout.
// Remove both env vars after the password has been reset.
func resetAdminIfRequested(ctx context.Context, q *store.Queries) error {
	email := os.Getenv("RESET_ADMIN_EMAIL")
	password := os.Getenv("RESET_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}

	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("reset admin: user %q not found", email)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	if err := q.UpdateUserPassword(ctx, store.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: hash,
	}); err != nil {
		return err
	}

	log.Printf("password reset on startup for: %s", email)
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
