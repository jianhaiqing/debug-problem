package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build delete_sql_limit.go
func main() {
	// const myDsn = "root:password@tcp(172.18.153.51:13336)/longrun"
	const myDsn = "root:password@tcp(mysql-master-5.gz.cvte.cn:3310)/seewo_mis_environment"
	db, err := sql.Open("mysql", myDsn)
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	delQuery := "DELETE from mis_env_report_data where id > 162786365912666115 and create_time < '2019-04-01 00:00:00' limit 2000;"
	for {
		res, err := db.Exec(delQuery)
		if err != nil {
			panic(err.Error())
		} else {
			count, err2 := res.RowsAffected()

			if err2 != nil {
				panic(err2.Error())
			} else {
				if count != 0 {
					fmt.Println("RowsAffected count:", count)
				} else {
					fmt.Println("Done: ", delQuery, "latest RowsAffected: ", count)
					break
				}
			}
		}
	}
}
