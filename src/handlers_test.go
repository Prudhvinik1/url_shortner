package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func setupRouterWithMockDB(t *testing.T, db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handler{DB: db}
	r := gin.New()
	r.GET("/:short_code", h.getLongUrl_service)
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
