package main

import (
	"flag"
	"fmt"
	"github.com/civet148/gotools/log"
	"github.com/civet148/sqlca"
	"github.com/civet148/sqlca/cmd/db2go/mysql"
	"github.com/civet148/sqlca/cmd/db2go/schema"
	"strings"
)

var argvUrl = flag.String("url", "", "mysql://root:123456@127.0.0.1:3306/test?charset=utf8")
var argvOutput = flag.String("out", ".", "output directory, default .")
var argvDatabase = flag.String("db", "", "export databases, like 'test,chat_db'")
var argvTables = flag.String("table", "", "export tables, eg. 'users,devices'")
var argvTags = flag.String("tag", "", "golang struct tag name, default json,db")
var argvPrefix = flag.String("prefix", "", "export file prefix")
var argvSuffix = flag.String("suffix", "", "export file suffix")
var argvPackage = flag.String("package", "", "export package name")
var argvWithout = flag.String("without", "", "exclude columns")
var argvReadOnly = flag.String("readonly", "", "read only columns")

func main() {

	//var err error
	var cmd schema.Commander
	log.Infof("argument: url [%v]", *argvUrl)
	log.Infof("argument: databases [%v]", *argvDatabase)
	log.Infof("argument: output [%v]", *argvOutput)
	log.Infof("argument: tag [%v]", *argvTags)
	log.Infof("argument: tables [%v]", *argvTables)
	log.Infof("argument: prefix [%v]", *argvPrefix)
	log.Infof("argument: suffix [%v]", *argvSuffix)
	log.Infof("argument: package [%v]", *argvPackage)
	log.Infof("argument: without [%v]", *argvWithout)
	log.Infof("argument: readonly [%v]", *argvReadOnly)

	if *argvUrl == "" {
		fmt.Println("need --url parameter")
		flag.Usage()
		return
	}

	if *argvTags != "" {
		cmd.Tags = trimSpaceSlice(strings.Split(*argvTags, ","))
	}
	if *argvReadOnly != "" {
		cmd.ReadOnly = trimSpaceSlice(strings.Split(*argvReadOnly, ","))
	}
	cmd.Prefix = *argvPackage
	cmd.Prefix = *argvPrefix
	cmd.Suffix = *argvSuffix
	cmd.OutDir = *argvOutput
	cmd.ConnUrl = *argvUrl
	cmd.PackageName = *argvPackage

	ui := sqlca.ParseUrl(*argvUrl)

	if *argvDatabase == "" {
		//use default database
		cmd.Databases = append(cmd.Databases, getDatabaseName(ui.Path))
	} else {
		//use input databases
		cmd.Databases = trimSpaceSlice(strings.Split(*argvDatabase, ","))
	}

	if *argvTables != "" {
		cmd.Tables = trimSpaceSlice(strings.Split(*argvTables, ","))
	}

	if *argvWithout != "" {
		cmd.Without = strings.Split(*argvWithout, ",")
	}

	cmd.Scheme = ui.Scheme
	cmd.Host = ui.Host
	cmd.User = ui.User
	cmd.Password = ui.Password

	switch cmd.Scheme {
	case "mysql":
		exportMysql(&cmd)
	case "postgres":
		exportPostgres(&cmd)
	}
}

func init() {
	flag.Parse()
}

func trimSpaceSlice(s []string) (ts []string) {
	for _, v := range s {
		ts = append(ts, strings.TrimSpace(v))
	}
	return
}

func getDatabaseName(strPath string) (strName string) {
	idx := strings.LastIndex(strPath, "/")
	if idx == -1 {
		return
	}
	return strPath[idx+1:]
}

func exportMysql(cmd *schema.Commander) {
	if err := mysql.Export(cmd); err != nil {
		log.Errorf("export mysql schema error [%v]", err.Error())
	}
}

func exportPostgres(cmd *schema.Commander) {

}
