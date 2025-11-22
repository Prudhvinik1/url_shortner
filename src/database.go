package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "urls"
)

func insertUser(db *sql.DB, username string) error {
	sqlStmnt := `INSERT INTO users (name) VALUES ($1)`
	_, err := db.Exec(sqlStmnt, username)
	return err
}

func insertUrl(db *sql.DB, short_code string, original_url string, is_alias bool, ttl int64, userid int64) error {
	sqlStmnt := `INSERT INTO urls (short_code,original_url,is_alias,ttl,user_id) VALUES ($1,$2,$3,$4,$5)`

	_, err := db.Exec(sqlStmnt, short_code, original_url, is_alias, ttl, userid)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func getLongUrl(db *sql.DB, short_code string) (string, error) {
	var long_url string
	var ttl int64

	sqlStmnt := `SELECT original_url, ttl FROM urls WHERE short_code = $1`

	row := db.QueryRow(sqlStmnt, short_code)

	err := row.Scan(&long_url, &ttl)

	if err != nil {
		fmt.Println("No rows were returned!")
		return "", err
	}

	//TODO: Implement TTL invalidation
	return long_url, nil

}

func initDB() *sql.DB {
	host := getEnv("POSTGRES_HOST", "localhost")
	portStr := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "")
	password := getEnv("POSTGRES_PASSWORD", "")
	dbname := getEnv("POSTGRES_DB", "urls")
	//sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("invalid POSTGRES_PORT: %v", err))
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	//sql.Open only validates if the credentials are correct
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	//db.Ping will force a connection establishment to the db server.
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("DB Connection Successful")
	return db
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
