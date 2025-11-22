package main

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInsertUrl_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs("abc123", "https://example.com", false, int64(0), int64(33)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := insertUrl(db, "abc123", "https://example.com", false, 0, 33); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInsertUrl_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs("dup", "https://x", true, int64(1), int64(33)).
		WillReturnError(sql.ErrConnDone) // simulate DB error

	if err := insertUrl(db, "dup", "https://x", true, 1, 33); err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetLongUrl_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"original_url", "ttl"}).
		AddRow("https://example.com", int64(0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original_url, ttl FROM urls WHERE short_code = $1`)).
		WithArgs("abc123").
		WillReturnRows(rows)

	long, err := getLongUrl(db, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if long != "https://example.com" {
		t.Fatalf("unexpected url: %s", long)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetLongUrl_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original_url, ttl FROM urls WHERE short_code = $1`)).
		WithArgs("nope").
		WillReturnError(sql.ErrNoRows)

	_, err = getLongUrl(db, "nope")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
