package main

import (
	"github.com/civet148/gotools/log"
	"github.com/civet148/sqlca/test/mssql"
	"github.com/civet148/sqlca/test/mysql"
	"github.com/civet148/sqlca/test/postgres"
)

func main() {
	log.Infof("------------------------------------------------------MYSQL------------------------------------------------------------")
	mysql.Benchmark()
	log.Infof("----------------------------------------------------SQLSERVER----------------------------------------------------------")
	mssql.Benchmark()
	log.Infof("----------------------------------------------------POSTGRES----------------------------------------------------------")
	postgres.Benchmark()

	//log.Info("%+v", log.Report()) //print function report
	log.Info("program exit...")
}
