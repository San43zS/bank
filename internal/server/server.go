package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"banking-platform/config"
	"banking-platform/internal/handler"
	"banking-platform/internal/middleware"
	"banking-platform/internal/service"
	"banking-platform/internal/storage"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router     *gin.Engine
	db         *storage.DB
	httpServer *http.Server
}

func NewServer(
	cfg *config.Config,
	db *storage.DB,
	authService service.IAuthService,
	accountService service.IAccountService,
	transactionService service.ITransactionService,
) *Server {
	router := gin.Default()
	if cfg != nil && cfg.RateLimitEnabled {
		router.Use(middleware.RateLimitMiddleware(cfg.RateLimitRPS, cfg.RateLimitBurst))
	}

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/me", middleware.AuthMiddleware(authService), authHandler.GetMe)
	}

	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.GET("/accounts", accountHandler.GetAccounts)
		protected.GET("/accounts/:id/balance", accountHandler.GetBalance)

		protected.POST("/transactions/transfer", transactionHandler.Transfer)
		protected.POST("/transactions/exchange", transactionHandler.Exchange)
		protected.GET("/transactions", transactionHandler.GetTransactions)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return &Server{
		router: router,
		db:     db,
	}
}

func (s *Server) Start(port string) error {
	addr := fmt.Sprintf(":%s", port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	log.Printf("Server starting on %s", addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Close() error {
	return s.db.Close()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
