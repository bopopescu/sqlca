package mysql

import (
	"fmt"
	"github.com/civet148/gotools/log"
	"github.com/civet148/sqlca"
	"github.com/civet148/sqlca/cmd/db2go/schema"
	"os"
	"strings"
)

const (
	IMPORT_SQLCA = `import "github.com/civet148/sqlca"`
)

/*
-- 查询数据库表名、引擎及注释
SELECT `TABLE_SCHEMA`, `TABLE_NAME`, `ENGINE`, `TABLE_COMMENT` FROM `INFORMATION_SCHEMA`.`TABLES`
WHERE `TABLE_SCHEMA`='accounts' AND (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')

-- 查询数据表字段名、字段类型及注释
SELECT `TABLE_NAME`, `COLUMN_NAME`, `DATA_TYPE`, `EXTRA`,  `COLUMN_KEY`, `COLUMN_COMMENT` FROM `INFORMATION_SCHEMA`.`COLUMNS`
WHERE `TABLE_SCHEMA` = 'accounts' AND `TABLE_NAME` = 'acc_3pl'
*/

func ExportGoStruct(cmd *schema.Commander, e *sqlca.Engine) (err error) {

	var strQuery string
	var tableSchemas []*schema.TableSchema

	var dbs, tables []string

	for _, v := range cmd.Databases {
		dbs = append(dbs, fmt.Sprintf("'%v'", v))
	}

	if len(dbs) == 0 {
		return fmt.Errorf("no database selected")
	}
	log.Infof("ready to export tables [%v]", cmd.Tables)
	for _, v := range cmd.Tables {
		tables = append(tables, fmt.Sprintf("'%v'", v))
	}

	if len(tables) == 0 {
		strQuery = fmt.Sprintf("SELECT `TABLE_SCHEMA`, `TABLE_NAME`, `ENGINE`, `TABLE_COMMENT` FROM `INFORMATION_SCHEMA`.`TABLES` "+
			"WHERE (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB') AND `TABLE_SCHEMA` IN (%v) ORDER BY TABLE_SCHEMA",
			strings.Join(dbs, ","))
	} else {
		strQuery = fmt.Sprintf("SELECT `TABLE_SCHEMA`, `TABLE_NAME`, `ENGINE`, `TABLE_COMMENT` FROM `INFORMATION_SCHEMA`.`TABLES` "+
			"WHERE (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB') AND `TABLE_SCHEMA` IN (%v) AND TABLE_NAME IN (%v) ORDER BY TABLE_SCHEMA",
			strings.Join(dbs, ","), strings.Join(tables, ","))
	}

	_, err = e.Model(&tableSchemas).QueryRaw(strQuery)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	return exportTableSchema(cmd, e, tableSchemas)
}

func exportTableSchema(cmd *schema.Commander, e *sqlca.Engine, tables []*schema.TableSchema) (err error) {

	for _, v := range tables {

		_, errStat := os.Stat(cmd.OutDir)
		if errStat != nil && os.IsNotExist(errStat) {

			log.Info("mkdir [%v]", cmd.OutDir)
			if err = os.Mkdir(cmd.OutDir, os.ModeDir); err != nil {
				log.Error("mkdir [%v] error (%v)", cmd.OutDir, err.Error())
				return
			}
		}

		v.OutDir = cmd.OutDir

		if cmd.PackageName == "" {
			//mkdir by output dir + scheme name
			cmd.PackageName = v.SchemeName
			if strings.LastIndex(cmd.OutDir, fmt.Sprintf("%v", os.PathSeparator)) == -1 {
				v.SchemeDir = fmt.Sprintf("%v/%v", cmd.OutDir, cmd.PackageName)
			} else {
				v.SchemeDir = fmt.Sprintf("%v%v", cmd.OutDir, cmd.PackageName)
			}
		} else {
			v.SchemeDir = fmt.Sprintf("%v/%v", cmd.OutDir, cmd.PackageName) //mkdir by package name
		}

		_, errStat = os.Stat(v.SchemeDir)

		if errStat != nil && os.IsNotExist(errStat) {

			log.Info("mkdir [%v]", v.SchemeDir)
			if err = os.Mkdir(v.SchemeDir, os.ModeDir); err != nil {
				log.Errorf("mkdir path name [%v] error (%v)", v.SchemeDir, err.Error())
				return
			}
		}

		var strPrefix, strSuffix string
		if cmd.Prefix != "" {
			strPrefix = fmt.Sprintf("%v_", cmd.Prefix)
		}
		if cmd.Suffix != "" {
			strSuffix = fmt.Sprintf("_%v", cmd.Suffix)
		}

		v.FileName = fmt.Sprintf("%v/%v%v%v.go", v.SchemeDir, strPrefix, v.TableName, strSuffix)
		if err = exportTableColumns(cmd, e, v); err != nil {
			return
		}
	}

	return
}

func exportTableColumns(cmd *schema.Commander, e *sqlca.Engine, table *schema.TableSchema) (err error) {

	var File *os.File
	File, err = os.OpenFile(table.FileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0)
	if err != nil {
		log.Errorf("open file [%v] error (%v)", table.FileName, err.Error())
		return
	}
	log.Infof("exporting table schema [%v] to file [%v]", table.TableName, table.FileName)

	var strHead, strContent string

	//write package name
	strHead += fmt.Sprintf("// Code generated by db2go. DO NOT EDIT.\n")
	strHead += fmt.Sprintf("package %v\n\n", cmd.PackageName)

	//var TableCols []schema.TableColumnDB
	//var TableColsGo []schema.TableColumnGo

	/*
	 SELECT `TABLE_NAME`, `COLUMN_NAME`, `DATA_TYPE`, `EXTRA`, `COLUMN_KEY`, `COLUMN_COMMENT` FROM `INFORMATION_SCHEMA`.`COLUMNS`
	 WHERE `TABLE_SCHEMA` = 'accounts' AND `TABLE_NAME` = 'users' ORDER BY ORDINAL_POSITION ASC
	*/
	e.Model(&table.Columns).QueryRaw("SELECT `TABLE_NAME`, `COLUMN_NAME`, `DATA_TYPE`, `EXTRA`, `COLUMN_KEY`, `COLUMN_COMMENT` FROM `INFORMATION_SCHEMA`.`COLUMNS` "+
		"WHERE `TABLE_SCHEMA` = '%v' AND `TABLE_NAME` = '%v' ORDER BY ORDINAL_POSITION ASC", table.SchemeName, table.TableName)

	//write table name in camel case naming
	strTableName := camelCaseConvert(table.TableName)
	table.TableComment = schema.ReplaceCRLF(table.TableComment)
	strContent += fmt.Sprintf("var TableName%v = \"%v\" //%v \n\n", strTableName, table.TableName, table.TableComment)

	table.StructName = fmt.Sprintf("%vDO", strTableName)

	if haveDecimal(table, table.Columns) {
		strHead += IMPORT_SQLCA + "\n\n" //根据数据库中是否存在decimal类型决定是否导入sqlca包
	}
	strContent += makeTableStructure(cmd, table)
	strContent += makeMethods(cmd, table)

	_, _ = File.WriteString(strHead + strContent)
	return
}

func haveDecimal(table *schema.TableSchema, TableCols []schema.TableColumn) (ok bool) {
	for _, v := range TableCols {
		_, ok = getGoColumnType(table.TableName, v.Name, v.DataType, v.Key, v.Extra, false)
		if ok {
			break
		}
	}
	return
}

func makeMethods(cmd *schema.Commander, table *schema.TableSchema) (strContent string) {

	for i, v := range table.Columns { //添加结构体成员Get/Set方法
		table.Columns[i].Comment = schema.ReplaceCRLF(v.Comment)
		if schema.IsInSlice(v.Name, cmd.Without) {
			continue
		}
		strColName := camelCaseConvert(v.Name)
		strColType, _ := getGoColumnType(table.TableName, v.Name, v.DataType, v.Key, v.Extra, cmd.DisableDecimal)
		strContent += schema.MakeGetter(table.StructName, strColName, strColType)
		if !schema.IsInSlice(v.Name, cmd.ReadOnly) {
			strContent += schema.MakeSetter(table.StructName, strColName, strColType)
		}
	}
	return
}

func makeTableStructure(cmd *schema.Commander, table *schema.TableSchema) (strContent string) {

	strContent += fmt.Sprintf("type %v struct { \n", table.StructName)

	for _, v := range table.Columns {

		if schema.IsInSlice(v.Name, cmd.Without) {
			continue
		}

		var tagValues []string
		var strColType, strColName string
		strColName = camelCaseConvert(v.Name)
		strColType, _ = getGoColumnType(table.TableName, v.Name, v.DataType, v.Key, v.Extra, cmd.DisableDecimal)

		if schema.IsInSlice(v.Name, cmd.ReadOnly) {
			tagValues = append(tagValues, fmt.Sprintf("%v:\"%v\"", sqlca.TAG_NAME_SQLCA, sqlca.SQLCA_TAG_VALUE_READ_ONLY))
		}
		for _, t := range cmd.Tags {
			tagValues = append(tagValues, fmt.Sprintf("%v:\"%v\"", t, v.Name))
		}
		//添加成员和标签
		strContent += schema.MakeTags(strColName, strColType, v.Name, v.Comment, strings.Join(tagValues, " "))

		v.GoName = strColName
		v.GoType = strColType
	}

	strContent += "}\n\n"

	return
}