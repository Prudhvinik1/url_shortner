package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

type urls struct {
	ID          string `json:"id"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	IsAlias     bool   `json:"isAlias"`
	TTL         int64  `json:"ttl"`
}

type Handler struct {
	DB *sql.DB
}

func (h *Handler) getLongUrl_service(c *gin.Context) {
	short_code := c.Param("short_code")
	long_url, err := getLongUrl(h.DB, short_code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, long_url)
}

func (h *Handler) healthz(c *gin.Context) {
	if err := h.DB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) postLongUrl(c *gin.Context) {
	if !authorizeWrite(c) {
		return
	}

	var newUrl urls

	//Call BindJson to bind the recieved JSON
	if err := c.BindJSON(&newUrl); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateOriginalURL(newUrl.OriginalURL); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	const maxAttempts = 5

	for attempts := 0; attempts < maxAttempts; attempts++ {
		candidate_shortcode, genErr := generateShortCode()
		if genErr != nil {
			fmt.Println("Generate ShortCode Error: ", genErr)
			continue
		}

		if err := insertUrl(h.DB, candidate_shortcode, newUrl.OriginalURL, newUrl.IsAlias, newUrl.TTL, 33); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.IndentedJSON(http.StatusAccepted, candidate_shortcode)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate unique shortcode"})
}

func authorizeWrite(c *gin.Context) bool {
	apiKey := os.Getenv("WRITE_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server not configured"})
		return false
	}

	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization"})
		return false
	}

	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token != apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return false
	}

	return true
}

func validateOriginalURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid url scheme")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func main() {
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	_ = godotenv.Load()

	db := initDB()
	defer db.Close() // Scheduled to run when main() returns

	h := &Handler{DB: db}

	router := gin.Default()
	router.GET("/:short_code", h.getLongUrl_service)
	router.POST("/urls", h.postLongUrl)
	router.GET("/healthz", h.healthz)

	port := getEnv("PORT", "8080")
	addr := ":" + port

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-stopCtx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
