package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlHandle struct {
	db       *sql.DB
	host     string
	port     int
	user     string
	password string
}

func NewMysql(host string, port int, user string, password string) *MysqlHandle {
	return &MysqlHandle{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (s *MysqlHandle) Connect() error {
	ds := fmt.Sprintf("%s:%s@tcp(%s:%d)/logist", s.user, s.password, s.host, s.port)
	db, err := sql.Open("mysql", ds)
	if err != nil {
		LogError("connect to db failed,  %s", err.Error())
		return err
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	s.db = db
	return nil
}

func (s *MysqlHandle) Exec(fmt string, args ...interface{}) error {
	_, err := s.db.Exec(fmt, args...)
	return err
}

func (s *MysqlHandle) ExecQuery(fmt string, args ...interface{}) *sql.Rows {
	rows, err := s.db.Query(fmt, args...)
	if err != nil {
		LogError("exec query failed, %s", err.Error())
		return nil
	}
	return rows

}
