package main

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

/************************
// use the following steps to reproduce deadlock of mysql5.7.24
1. "FLUSH LOGS;",
2. "PURGE BINARY LOGS BEFORE '2038-01-19';",
3. "SELECT * FROM performance_schema.session_variables WHERE VARIABLE_NAME LIKE 'binlog_transaction_dependency_tracking';",
4. "SHOW BINARY LOGS;"
*****************************/

func TestConnect(t *testing.T) {
	db, err := sql.Open("mysql", MysqlDNS)

	//var sql[4] string = { FLUSHLOGS, PURGEBIN, SELECTTRACT, SHOWBIN }

	t.Logf("hello world")

	if err != nil {
		t.Fatalf("%s failed to connect", MysqlDNS)
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		t.Fatalf("%s failed to ping", MysqlDNS)
	}

	ConnectMysql(db, FLUSHLOGS, 10)
}
