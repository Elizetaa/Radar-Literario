package database

import (
	"context"
	"database/sql"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"radar-literario/backend/internal/config"
	"strconv"
	"time"
)

func DSN(c config.Config) string {
	d := mysqlDriver.NewConfig()
	d.User = c.DBUser
	d.Passwd = c.DBPassword
	d.Net = "tcp"
	d.Addr = net.JoinHostPort(c.DBHost, strconv.Itoa(c.DBPort))
	d.DBName = c.DBName
	d.ParseTime = true
	d.Loc = time.UTC
	d.Timeout = c.DBTimeout
	d.ReadTimeout = c.DBTimeout
	d.WriteTimeout = c.DBTimeout
	d.Collation = "utf8mb4_unicode_ci"
	d.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
	return d.FormatDSN()
}
func Open(c config.Config) (*gorm.DB, *sql.DB, error) {
	db, e := gorm.Open(mysql.Open(DSN(c)), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true})
	if e != nil {
		return nil, nil, fmt.Errorf("falha ao configurar conexão MariaDB")
	}
	pool, e := db.DB()
	if e != nil {
		return nil, nil, e
	}
	pool.SetMaxOpenConns(c.DBMaxOpen)
	pool.SetMaxIdleConns(c.DBMaxIdle)
	pool.SetConnMaxLifetime(c.DBLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), c.DBTimeout)
	defer cancel()
	if e = pool.PingContext(ctx); e != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("MariaDB indisponível; confira host, porta, banco e credenciais locais")
	}
	return db, pool, nil
}
