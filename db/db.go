package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type status int8

const (
	maintanance status = iota
	available
	booked

	dbTimeout = time.Second * 3
	size      = 10

	dateFormat = "2006-01-02 15:04:05"
)

var (
	ErrNilQueryRowContext = errors.New("no data")
	ErrAlreadyUnparked    = errors.New("already unparked")
)

type DB struct {
	dbConn *sql.DB
}

func (d *DB) Close() error {
	return d.dbConn.Close()
}

func NewDB(dsn string) (*DB, error) {
	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	dbConn.SetConnMaxLifetime(time.Minute * 3)
	dbConn.SetMaxOpenConns(10)
	dbConn.SetMaxIdleConns(10)

	return &DB{
		dbConn: dbConn,
	}, nil
}

func getOffset(p int) int {
	return (p - 1) * size
}

func (s status) value() string {
	switch s {
	case maintanance:
		return "IN_MAINTANANCE"
	case available:
		return "AVAILABLE"
	case booked:
		return "BOOKED"
	}

	panic("NO_MATCH_FOUND")
}
