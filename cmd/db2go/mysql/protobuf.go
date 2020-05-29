package mysql

import (
	"fmt"
	"github.com/civet148/gotools/log"
	"github.com/civet148/sqlca"
	"github.com/civet148/sqlca/cmd/db2go/schema"
	"os"
)

func ExportProtobuf(cmd *schema.Commander, e *sqlca.Engine) (err error) {

	var schemas []*schema.TableSchema
	if schemas, err = queryTables(cmd, e); err != nil {
		log.Errorf(err.Error())
		return
	}
	for _, v := range schemas {
		if err = queryTableColumns(cmd, e, v); err != nil {
			log.Error(err.Error())
			return
		}
		var file *os.File
		if file, err = schema.CreateOutputFile(cmd, v, "proto"); err != nil {
			log.Error(err.Error())
			return
		}

		strHead := makeProtoHead(cmd, v)
		strBody := makeProtoBody(cmd, v)
		file.WriteString(strHead + strBody)
	}
	return
}

func makeProtoHead(cmd *schema.Commander, table *schema.TableSchema) (strContent string) {

	strContent += "syntax = \"proto3\";\n"
	strContent += fmt.Sprintf("package %v;\n\n", cmd.PackageName)

	if len(cmd.GogoOptions) > 0 {
		strContent += schema.IMPORT_GOGO_PROTO + "\n"
		//strContent += schema.IMPORT_GOOGOLE_PROTOBUF + "\n"
	}
	strContent += "\n"
	for _, v := range cmd.GogoOptions {
		strContent += fmt.Sprintf("option %v;\n", v)
	}
	strContent += "\n"
	return
}

func makeProtoBody(cmd *schema.Commander, table *schema.TableSchema) (strContent string) {

	strTableName := schema.CamelCaseConvert(table.TableName)
	strContent += fmt.Sprintf("message %vDO {\n", strTableName)
	for i, v := range table.Columns {
		no := i + 1
		strColName := v.Name
		strColType := getProtoColumnType(table.TableName, v.Name, v.DataType, v.Key, v.Extra, true)
		strContent += fmt.Sprintf("	%-10s %-22s = %-2d; //%v\n", strColType, strColName, no, v.Comment)
	}
	strContent += "}\n\n"
	return
}
