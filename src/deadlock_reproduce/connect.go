package main

import (
	"database/sql"
	"fmt"

	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 复现 数据库死锁的几条SQL ，还有数据库连接配置
const (
	FLUSHLOGS   = "FLUSH LOGS;"
	PURGEBIN    = "PURGE BINARY LOGS BEFORE '2038-01-19';"
	SELECTTRACT = "SELECT * FROM performance_schema.session_variables WHERE VARIABLE_NAME LIKE 'binlog_transaction_dependency_tracking';"
	SHOWBIN     = "SHOW BINARY LOGS;"
	MysqlDNS    = "root:password@tcp(10.21.17.74:23306)/"
)

// ConnectMysql used in the loop
// db connection pool for mysql
// querySQL used to query
// queryCount count to execute querySQL, 0 means infinite
func ConnectMysql(db *sql.DB, querySQL string, queryCount int) {
	defer db.Close()

	if queryCount < 0 {
		fmt.Printf("queryCount is %d, must be larger than zero", queryCount)
		return
	}

	// Open doesn't open a connection. Validate DSN data:
	err := db.Ping()
	if err != nil {
	}
	bInfinite := false
	if queryCount == 0 {
		bInfinite = true
	}
	for {
		if queryCount < 0 {
			break
		}
		// Execute the query
		rows, err := db.Query(querySQL)
		fmt.Printf("%s, %s in goroutine.", time.Now(), querySQL)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Make a slice for the values
		values := make([]sql.RawBytes, len(columns))

		// rows.Scan wants '[]interface{}' as an argument, so we must copy the
		// references into such a slice
		// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Fetch rows
		for rows.Next() {
			// get RawBytes from data
			err = rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}

			// Now do something with the data.
			// Here we just print each column as a string.
			var value string
			for i, col := range values {
				// Here we can check if the value is nil (NULL value)
				if col == nil {
					value = "NULL"
				} else {
					value = string(col)
				}
				fmt.Println(columns[i], ": ", value)
			}
			fmt.Println("-----------------------------------")
		}
		if err = rows.Err(); err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		if !bInfinite {
			queryCount--
		}

		time.Sleep(100 * time.Microsecond)

	}

}

func main() {

	//var sql[4] string = { FLUSHLOGS, PURGEBIN, SELECTTRACT, SHOWBIN }
	var sqls [4]string
	sqls[0] = FLUSHLOGS
	sqls[1] = PURGEBIN
	sqls[2] = SELECTTRACT
	sqls[3] = SHOWBIN

	for _, queryStr := range sqls {
		db, err := sql.Open("mysql", MysqlDNS)
		// Open doesn't open a connection. Validate DSN data:
		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}

		go ConnectMysql(db, queryStr, 0)
	}

	fmt.Printf("main func should be hold the thread.")
	for {
		fmt.Printf("%s, main func should be hold the thread.", time.Now())
		time.Sleep(1000 * time.Millisecond)
	}

}
