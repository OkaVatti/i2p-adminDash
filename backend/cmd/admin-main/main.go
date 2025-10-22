package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware" // keep logger/recover
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"i2p-adminDash/m/config"
	"i2p-adminDash/m/internal/auth"
	"i2p-adminDash/m/internal/i2psnark"
	"i2p-adminDash/m/internal/models"
	"i2p-adminDash/m/internal/routerproxy"
	"i2p-adminDash/m/internal/ws"
)

func main() {
	cfg := config.Load()

	// Redis client (for caching / pubsub)
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// create Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Basic health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Setup API group
	api := e.Group("/api")

	// Unauthenticated endpoints
	api.POST("/setup", func(c echo.Context) error {
		type req struct {
			Username  string `json:"username"`
			Email     string `json:"email"`
			Sha3pwHex string `json:"sha3sha3"`
		}
		var r req
		if err := c.Bind(&r); err != nil {
			return echo.ErrBadRequest
		}
		// Prevent setup if encrypted DB already present
		if auth.EncryptedDBExists() {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "already_setup"})
		}
		phc, raw, err := auth.HashToPHC(r.Sha3pwHex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash_failed"})
		}
		// Encrypt DSN using raw (derived key)
		if err := auth.EncryptDBCreds(cfg.PostgresDSN, raw); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "encrypt_db_failed"})
		}
		// verify decrypt
		dsn, err := auth.DecryptDBCreds(raw)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "decrypt_verify_failed"})
		}
		// Open DB and create user record with PHC string
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db_connect_failed"})
		}
		if err := gdb.AutoMigrate(&models.User{}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "migrate_failed"})
		}
		u := models.User{
			Username:  r.Username,
			Email:     r.Email,
			ArgonPHC:  phc,
			CreatedAt: time.Now(),
		}
		if err := gdb.Create(&u).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "create_user_failed"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "setup_complete"})
	})

	api.POST("/login", func(c echo.Context) error {
		type req struct {
			Username string `json:"username"`
			Sha3hex  string `json:"sha3"`
		}
		var r req
		if err := c.Bind(&r); err != nil {
			return echo.ErrBadRequest
		}
		if !auth.EncryptedDBExists() {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "not_setup"})
		}

		// Fetch user PHC using environment DSN (pragmatic approach as explained earlier)
		gdb, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db_access_failed"})
		}
		var user models.User
		if err := gdb.Where("username = ?", r.Username).First(&user).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user_not_found"})
		}
		derived, ok, err := auth.VerifyPHC(r.Sha3hex, user.ArgonPHC)
		if err != nil || !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		}
		// Decrypt the encrypted DSN using derived key and open production DB
		dsn, err := auth.DecryptDBCreds(derived)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db_decrypt_failed"})
		}
		prodDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db_open_failed"})
		}
		var u models.User
		_ = prodDB.Where("username = ?", r.Username).First(&u) // best-effort sync
		// create JWT
		token, err := auth.CreateJWT(u.ID, cfg.JWTSecret, time.Minute*45)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "jwt_failed"})
		}
		return c.JSON(http.StatusOK, map[string]string{"token": token})
	})

	// Protected routes group with custom JWT middleware
	apiAuth := api.Group("")
	apiAuth.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check Authorization header: "Bearer <token>"
			authz := c.Request().Header.Get("Authorization")
			if authz == "" {
				// allow token via query param for WebSocket handshake (legacy)
				authz = c.QueryParam("token")
				if authz == "" {
					return echo.ErrUnauthorized
				}
			}
			if strings.HasPrefix(authz, "Bearer ") {
				authz = strings.TrimPrefix(authz, "Bearer ")
			}
			// verify
			uid, err := auth.VerifyJWT(authz, cfg.JWTSecret)
			if err != nil {
				return echo.ErrUnauthorized
			}
			// store uid in context for handlers
			c.Set("uid", uid)
			return next(c)
		}
	})

	// Example: proxy to router console (protected)
	apiAuth.GET("/router/stats", func(c echo.Context) error {
		stats, err := routerproxy.FetchRouterStats(cfg.RouterConsoleURL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "router_fetch_failed"})
		}
		return c.JSON(http.StatusOK, stats)
	})

	apiAuth.POST("/i2psnark/command", func(c echo.Context) error {
		type req struct {
			Cmd string `json:"cmd"`
			Arg string `json:"arg"`
		}
		var r req
		if err := c.Bind(&r); err != nil {
			return echo.ErrBadRequest
		}
		out, err := i2psnark.RunCommand(r.Cmd, r.Arg)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "i2psnark_failed", "detail": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"out": out})
	})

	// websocket hub
	hub := ws.NewHub(rdb)
	go hub.Run()
	e.GET("/ws", func(c echo.Context) error {
		// token may be passed as query param (ws client does this)
		token := c.QueryParam("token")
		if token == "" {
			// or Authorization header Bearer
			authz := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(authz, "Bearer ") {
				token = strings.TrimPrefix(authz, "Bearer ")
			}
		}
		if token == "" {
			return c.NoContent(http.StatusUnauthorized)
		}
		uid, err := auth.VerifyJWT(token, cfg.JWTSecret)
		if err != nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		uidStr := fmt.Sprintf("%d", uid)
		ws.ServeWS(hub, c.Response(), c.Request(), uidStr)
		return nil
	})

	// static frontend
	// use SvelteKit static build output or dev server in front
	e.Static("/", "../../web/public")

	// create server
	addr := fmt.Sprintf("%s:%d", cfg.BindHost, cfg.BindPort)
	log.Printf("listening on %s\n", addr)
	// start echo
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("shutting down: %v", err)
	}
	// graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.Shutdown(ctx)
}
