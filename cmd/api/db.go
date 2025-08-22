package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

type postgresDbConn struct {
	dbConnStr          string
	maxOpenConnections int
	maxIdleConnections int
	connMaxLifeTime    time.Duration
	connMaxIdleTime    time.Duration
}

func newPostgresDbConn(dbConnStr string, maxOpenConnections int, maxIdleConnections int, connMaxLifeTime time.Duration, connMaxIdleTime time.Duration) *postgresDbConn {
	return &postgresDbConn{
		dbConnStr:          dbConnStr,
		maxOpenConnections: maxOpenConnections,
		maxIdleConnections: maxIdleConnections,
		connMaxLifeTime:    connMaxLifeTime,
		connMaxIdleTime:    connMaxIdleTime,
	}
}

func (p *postgresDbConn) connect() (*sqlx.DB, error) {

	db, err := sqlx.Open("postgres", p.dbConnStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(p.maxOpenConnections)
	db.SetMaxIdleConns(p.maxIdleConnections)
	db.SetConnMaxLifetime(p.connMaxLifeTime)
	db.SetConnMaxIdleTime(p.connMaxIdleTime)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
