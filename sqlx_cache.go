package sqlca

import (
	"fmt"
	"strings"
)

const (
	TAG_NAME_DB    = "db"
	TAG_NAME_JSON  = "json"
	TAG_NAME_REDIS = "redis"
)

type AdapterType int

const (
	DEFAULT_PRIMARY_KEY_NAME = "id"
)

const (
	AdapterSqlx_MySQL      AdapterType = 1  //sqlx: mysql
	AdapterSqlx_Postgres   AdapterType = 2  //sqlx: postgresql
	AdapterSqlx_Sqlite     AdapterType = 3  //sqlx: sqlite
	AdapterSqlx_Mssql      AdapterType = 4  //sqlx: mssql server
	AdapterCache_Redis     AdapterType = 11 //cache: redis
	AdapterCache_Memcached AdapterType = 12 //cache: memcached
	AdapterCache_Memory    AdapterType = 13 //cache: memory
	AdapterCache_File      AdapterType = 14 //cache: file
)

func (a AdapterType) GoString() string {
	return a.String()
}

func (a AdapterType) String() string {

	switch a {
	case AdapterSqlx_MySQL:
		return "AdapterSqlx_MySQL"
	case AdapterSqlx_Postgres:
		return "AdapterSqlx_Postgres"
	case AdapterSqlx_Sqlite:
		return "AdapterSqlx_Sqlite"
	case AdapterSqlx_Mssql:
		return "AdapterSqlx_Mssql"
	case AdapterCache_Redis:
		return "AdapterCache_Redis"
	case AdapterCache_Memcached:
		return "AdapterCache_Memcached"
	case AdapterCache_Memory:
		return "AdapterCache_Memory"
	case AdapterCache_File:
		return "AdapterCache_File"
	default:
	}
	return "Adapter_Unknown"
}

func (a AdapterType) Schema() string {
	switch a {
	case AdapterSqlx_MySQL:
		return "mysql"
	case AdapterSqlx_Postgres:
		return "postges"
	case AdapterSqlx_Sqlite:
		return "sqlite"
	case AdapterSqlx_Mssql:
		return "mssql"
	case AdapterCache_Redis:
		return "redis"
	case AdapterCache_Memcached:
		return "memcached"
	case AdapterCache_Memory:
		return "memory"
	case AdapterCache_File:
		return "file"
	default:
	}
	return "unknown"
}

type OperType int

const (
	OperType_Query  OperType = 1 // query sql
	OperType_Update OperType = 2 // update sql
	OperType_Insert OperType = 3 // insert sql
	OperType_Upsert OperType = 4 // insert or update sql
	OperType_Tx     OperType = 5 // transaction sql

)

func (o OperType) GoString() string {
	return o.String()
}

func (o OperType) String() string {
	switch o {
	case OperType_Query:
		return "OperType_Query"
	case OperType_Update:
		return "OperType_Update"
	case OperType_Insert:
		return "OperType_Insert"
	case OperType_Upsert:
		return "OperType_Upsert"
	case OperType_Tx:
		return "OperType_Tx"
	}
	return "OperType_Unknown"
}

type ModeType int

const (
	ModeType_ORM = 1
	ModeType_Raw = 2
)

func (m ModeType) GoString() string {
	return m.String()
}

func (m ModeType) String() string {
	switch m {
	case ModeType_ORM:
		return "ModeType_ORM"
	case ModeType_Raw:
		return "ModeType_Raw"
	}
	return "ModeType_Unknown"
}

// clone engine
func (e *Engine) clone(model interface{}) *Engine {

	dict := Struct(model).ToMap(TAG_NAME_DB)
	return &Engine{
		db:        e.db,
		cache:     e.cache,
		debug:     e.debug,
		model:     model,
		dict:      dict,
		strPkName: e.strPkName,
	}
}

func (e *Engine) checkModel() bool {

	if e.model == nil {
		e.panic("orm model is nil, please call Model() method before query or update")
		return false
	}
	return true
}

func (e *Engine) getTableName() string {
	return e.strTableName
}

func (e *Engine) setTableName(strName string) {
	e.strTableName = strName
}

func (e *Engine) getPkName() string {
	return e.strPkName
}

func (e *Engine) setPkName(strName string) {
	e.strPkName = strName
}

func (e *Engine) getPkValue() string {
	return e.strPkValue
}

func (e *Engine) setPkValue(strValue string) {
	e.strPkValue = strValue
}

func (e *Engine) getColumns() []string {
	return e.strColumns
}

func (e *Engine) setColumns(strColumns ...string) {
	e.strColumns = strColumns
}

func (e *Engine) getWhere() string {
	return e.strWhere
}

func (e *Engine) setWhere(strWhere string) {
	e.strWhere = strWhere
}

func (e *Engine) getModeType() ModeType {
	return e.modeType
}

func (e *Engine) setModeType(modeType ModeType) {
	e.modeType = modeType
}

func (e *Engine) getOperType() OperType {
	return e.operType
}

func (e *Engine) setOperType(operType OperType) {
	e.operType = operType
}

// get data base driver name and data source name
func (e *Engine) getConnUrl(adapterType AdapterType, strUrl string) (strScheme, strDSN string) {
	//TODO @libin parse connect url for database
	strScheme = adapterType.Schema()
	return strScheme, strUrl
}

func (e *Engine) getForwardQuote() (strSlash string) {
	switch e.adapterSqlx {
	case AdapterSqlx_MySQL:
		return "`"
	case AdapterSqlx_Postgres:
		return ""
	case AdapterSqlx_Sqlite:
		return ""
	case AdapterSqlx_Mssql:
		return ""
	}
	return
}

func (e *Engine) getBackQuote() (strSlash string) {
	switch e.adapterSqlx {
	case AdapterSqlx_MySQL:
		return "`"
	case AdapterSqlx_Postgres:
		return ""
	case AdapterSqlx_Sqlite:
		return ""
	case AdapterSqlx_Mssql:
		return ""
	}
	return
}

func (e *Engine) getOnDuplicateForwardKey() (strKey string) {
	switch e.adapterSqlx {
	case AdapterSqlx_MySQL:
		return " ON DUPLICATE KEY "
	case AdapterSqlx_Postgres:
		return " ON CONFLICT ( "
	case AdapterSqlx_Sqlite:
		return " "
	case AdapterSqlx_Mssql:
		return " "
	}
	return
}

func (e *Engine) getOnDuplicateBackKey() (strKey string) {
	switch e.adapterSqlx {
	case AdapterSqlx_MySQL:
		return " UPDATE "
	case AdapterSqlx_Postgres:
		return " ) DO UPDATE SET "
	case AdapterSqlx_Sqlite:
		return " "
	case AdapterSqlx_Mssql:
		return " "
	}
	return
}

func (e *Engine) getOnDuplicateUpdate() (strUpdate string) {
	// TODO @libin insert into table(...) value(...) on duplicate...
	return
}

func (e *Engine) isColumnSelected(strCol string, strExcepts ...string) bool {

	if len(e.strColumns) == 0 {
		return true
	}

	for _, v := range strExcepts {
		if v == strCol {
			return true
		}
	}

	for _, v := range e.strColumns {
		if v == strCol {
			return true
		}
	}
	return false
}

func (e *Engine) getQuoteConflicts() (strQuoteConflicts string) {
	var cols []string

	for _, v := range e.strConflicts {

		if e.isColumnSelected(v) {
			c := fmt.Sprintf("%v%v%v", e.getForwardQuote(), v, e.getBackQuote()) // postgresql conflict column name format to `id`,...
			cols = append(cols, c)
		}
	}

	if len(cols) > 0 {
		strQuoteConflicts = strings.Join(cols, ",")
	}
	return
}

func (e *Engine) getQuoteColumns(strExcepts ...string) (strQuoteColumns string) {
	var cols []string

	for _, v := range e.dict {

		if e.isColumnSelected(v, strExcepts...) {
			c := fmt.Sprintf("%v%v%v", e.getForwardQuote(), v, e.getBackQuote()) // column name format to `id`,...
			cols = append(cols, c)
		}
	}
	strQuoteColumns = strings.Join(cols, ",")
	return
}

func (e *Engine) getColonValues(strExcepts ...string) (strQuoteValues string) {
	var cols []string
	for _, v := range e.dict {
		if e.isColumnSelected(v, strExcepts...) {
			c := fmt.Sprintf(":%v", v) // column value format to :id,...
			cols = append(cols, c)
		}
	}
	strQuoteValues = strings.Join(cols, ",")
	return
}

func (e *Engine) makeSqlxString() (strSqlx string) {
	assert(e.getModeType() == ModeType_ORM, "not a orm mode")
	//TODO: @libin make SQL query string (mysql)

	switch e.operType {
	case OperType_Query:
	case OperType_Update:
	case OperType_Insert:
	case OperType_Upsert:
	case OperType_Tx:
	default:
		assert(false, "operation illegal")
	}

	if e.debug {
		e.debugf("SqlxString: %s", strSqlx)
	}
	return
}

func (e *Engine) makeSqlxQuery() (strSqlx string) {

	return
}

func (e *Engine) makeSqlxUpdate() (strSqlx string) {

	return
}

//"insert into
// phone_call_sessions(`access_hash`, `admin_id`, `participant_id`, `admin_auth_key_id`, `participant_auth_key_id`, `random_id`, `admin_protocol`,
//                     `participant_protocol`, `g_a_hash`, `g_a`, `g_b`, `key_fingerprint`, `connections`, `admin_debug_data`, `participant_debug_data`, `date`, `state`)
// values (:access_hash, :admin_id, :participant_id, :admin_auth_key_id, :participant_auth_key_id, :random_id, :admin_protocol,
//         :participant_protocol, :g_a_hash, :g_a, :g_b, :key_fingerprint, :connections, :admin_debug_data, :participant_debug_data, :date, :state)"
func (e *Engine) makeSqlxInsert() (strSqlx string) {

	strSqlx = fmt.Sprintf("insert into %v (%v) values (%v)", e.strTableName, e.getQuoteColumns(), e.getColonValues())

	return
}

//"insert into
// phone_call_sessions(access_hash, admin_id, participant_id, admin_auth_key_id, participant_auth_key_id, random_id, admin_protocol, participant_protocol, g_a_hash, g_a, g_b, key_fingerprint, connections, admin_debug_data, participant_debug_data, `date`, state)
// values (:access_hash, :admin_id, :participant_id, :admin_auth_key_id, :participant_auth_key_id, :random_id, :admin_protocol,
//         :participant_protocol, :g_a_hash, :g_a, :g_b, :key_fingerprint, :connections, :admin_debug_data, :participant_debug_data, :date, :state)
// on duplicate key update id = last_insert_id(id), date='1582890480'"
func (e *Engine) makeSqlxUpsert() (strSqlx string) {

	return
}

func (e *Engine) makeSqlxTx() (strSqlx string) {

	return
}
