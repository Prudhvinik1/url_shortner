package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func setupRouterWithMockDB(t *testing.T, db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handler{DB: db}
	r := gin.New()
	r.GET("/:short_code", h.getLongUrl_service)
	r.POST("/urls", h.postLongUrl)
	r.GET("/healthz", h.healthz)
	return r
}

func TestGetLongUrlService_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original_url, ttl FROM urls WHERE short_code = $1`)).
		WithArgs("miss").
		WillReturnError(sql.ErrNoRows)

	router := setupRouterWithMockDB(t, db)

	req := httptest.NewRequest("GET", "/miss", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetLongUrlService_Found_Redirect(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"original_url", "ttl"}).
		AddRow("https://example.com", int64(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original_url, ttl FROM urls WHERE short_code = $1`)).
		WithArgs("ok").
		WillReturnRows(rows)

	router := setupRouterWithMockDB(t, db)

	req := httptest.NewRequest("GET", "/ok", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect status, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Fatalf("unexpected Location header: %s", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostLongUrl_Unauthorized(t *testing.T) {
	t.Setenv("WRITE_API_KEY", "secret")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	router := setupRouterWithMockDB(t, db)

	body := `{"original_url":"https://example.com","isAlias":false,"ttl":0}`
	req := httptest.NewRequest("POST", "/urls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPostLongUrl_InvalidURL(t *testing.T) {
	t.Setenv("WRITE_API_KEY", "secret")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	router := setupRouterWithMockDB(t, db)

	body := `{"original_url":"ftp://example.com","isAlias":false,"ttl":0}`
	req := httptest.NewRequest("POST", "/urls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPostLongUrl_Success(t *testing.T) {
	t.Setenv("WRITE_API_KEY", "secret")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(sqlmock.AnyArg(), "https://example.com", false, int64(0), int64(33)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := setupRouterWithMockDB(t, db)

	body := `{"original_url":"https://example.com","isAlias":false,"ttl":0}`
	req := httptest.NewRequest("POST", "/urls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostLongUrl_RetryOnUniqueViolation(t *testing.T) {
	t.Setenv("WRITE_API_KEY", "secret")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(sqlmock.AnyArg(), "https://example.com", false, int64(0), int64(33)).
		WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(sqlmock.AnyArg(), "https://example.com", false, int64(0), int64(33)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	router := setupRouterWithMockDB(t, db)

	body := `{"original_url":"https://example.com","isAlias":false,"ttl":0}`
	req := httptest.NewRequest("POST", "/urls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHealthz_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(nil)

	router := setupRouterWithMockDB(t, db)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
