package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Query struct {
	user     string
	password string
	port     int
	host     string
	database string
	sql      string
	limit    int
	interval int
	timeout  time.Duration
}

func FormatDSN(query Query) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", query.user, query.password, query.host, query.port, query.database)
}

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build deletesql.go
func main() {
	var query Query
	str := os.Args
	fmt.Println(str)
	flag.StringVar(&query.host, "host", "localhost", "MySQL host")
	flag.StringVar(&query.user, "user", "root", "MySQL user")
	flag.StringVar(&query.database, "database", "test", "MySQL database name")
	flag.IntVar(&query.port, "port", 3306, "MySQL port")
	flag.StringVar(&query.password, "password", "", "MySQL password")
	flag.StringVar(&query.sql, "sql", "", "delete sql to be executed")
	flag.IntVar(&query.limit, "limit", 1000, "MySQL limit 1000")
	flag.IntVar(&query.interval, "interval", 1000, "MySQL run loop in interval millisecond")
	flag.Parse()
	myDsn := FormatDSN(query)
	db, err := sql.Open("mysql", myDsn)
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	query.sql = strings.Trim(query.sql, ";")
	delQuery := fmt.Sprintf("%s limit %d", query.sql, query.limit)
	var totalCount int64

	for {
		res, err := db.Exec(delQuery)
		if err != nil {
			panic(err.Error())
		} else {
			count, err2 := res.RowsAffected()
			totalCount += count

			if err2 != nil {
				panic(err2.Error())
			} else {
				if count != 0 {
					log.Printf("RowsAffected count: %d", count)
				} else {
					log.Printf("Done: %s, latest RowsAffected: %d", delQuery, count)
					break
				}
			}
			log.Printf("Total rows affected: %d", totalCount)
		}
		if query.interval > 0 {
			time.Sleep(time.Duration(query.interval) * time.Millisecond)
		}
	}
}
