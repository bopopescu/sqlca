package sqlca

import (
	"database/sql"
	"fmt"
	"github.com/civet148/gotools/log"
	"github.com/civet148/redigogo"
	_ "github.com/go-sql-driver/mysql" //mysql golang driver
	"github.com/jmoiron/sqlx"          //sqlx package
	_ "github.com/lib/pq"              //postgres golang driver
	_ "github.com/mattn/go-adodb"      //mssql golang driver
	_ "github.com/mattn/go-sqlite3"    //sqlite3 golang driver
	"strings"
)

type Engine struct {
	db              *sqlx.DB               // sqlx instance
	cache           redigogo.Cache         // redis cache instance
	adapterSqlx     AdapterType            // what's adapter of sqlx
	adapterCache    AdapterType            // what's adapter of cache
	modelType       ModelType              // model type
	operType        OperType               // operation type
	expireTime      int                    // cache expire time of seconds
	bUseCache       bool                   // can update to cache or read from cache? (true=yes false=no)
	bCacheFirst     bool                   // cache first or database first (true=cache first; false=db first)
	debug           bool                   // debug mode [on/off]
	model           interface{}            // data model [struct object or struct slice]
	dict            map[string]interface{} // data model db dictionary
	strTableName    string                 // table name
	strPkName       string                 // primary key of table, default 'id'
	strPkValue      string                 // primary key's value
	strWhere        string                 // where condition to query or update
	strLimit        string                 // limit
	strOffset       string                 // offset (only for postgres)
	strAscOrDesc    string                 // order by ... [asc|desc]
	selectColumns   []string               // columns to query: select
	conflictColumns []string               // conflict key on duplicate set (just for postgresql)
	orderByColumns  []string               // order by columns
	groupByColumns  []string               // group by columns
	cacheIndexes    []tableIndex           // index read or write cache
}

func init() {
	log.SetLevel(log.LEVEL_INFO)
}

func NewEngine(debug bool) *Engine {

	return &Engine{
		debug:      debug,
		strPkName:  DEFAULT_PRIMARY_KEY_NAME,
		expireTime: DEFAULT_CAHCE_EXPIRE_SECONDS,
	}
}

// get data base driver name and data source name
func (e *Engine) getConnConfig(adapterType AdapterType, strUrl string) (strDriverName, strDSN string) {

	strDriverName = adapterType.DriverName()
	switch adapterType {
	case AdapterSqlx_MySQL:
		return strDriverName, e.parseMysqlUrl(strUrl)
	case AdapterSqlx_Postgres:
		return strDriverName, e.parsePostgresUrl(strUrl)
	case AdapterSqlx_Sqlite:
		return strDriverName, e.parseSqliteUrl(strUrl)
	case AdapterSqlx_Mssql:
		return strDriverName, e.parseMssqlUrl(strUrl)
	case AdapterCache_Redis:
		return strDriverName, e.parseRedisUrl(strUrl)
	}
	return strDriverName, strUrl
}

// open a sqlx database or cache connection
// strUrl:
//
//  1. data source name
//
// 	   [mysql]    Open(AdapterSqlx_MySQL,    "mysql://root:123456@127.0.0.1:3306/mydb?charset=utf8mb4")
// 	   [postgres] Open(AdapterSqlx_Postgres, "postgres://root:123456@127.0.0.1:5432/mydb?sslmode=disable")
// 	   [sqlite]   Open(AdapterSqlx_Sqlite,   "sqlite:///var/lib/my.db")
// 	   [mssql]    Open(AdapterSqlx_Mssql,    "mssql://sa:123456@127.0.0.1:1433/mydb?instance=&windows=false")
//
//  2. cache config
//     [redis-alone]    Open(AdapterTypeCache_Redis,    "redis://123456@127.0.0.1:6379/cluster?db=0")
//     [redis-cluster]  Open(AdapterTypeCache_Redis,    "redis://123456@127.0.0.1:6379/cluster?db=0&replicate=127.0.0.1:6380,127.0.0.1:6381")
//
// expireSeconds cache data expire seconds, just for AdapterTypeCache_XXX
func (e *Engine) Open(adapterType AdapterType, strUrl string, expireSeconds ...int) *Engine {

	var err error
	strDriverName, strConfig := e.getConnConfig(adapterType, strUrl)
	switch adapterType {
	case AdapterSqlx_MySQL, AdapterSqlx_Postgres, AdapterSqlx_Sqlite, AdapterSqlx_Mssql:
		if e.db, err = sqlx.Open(strDriverName, strConfig); err != nil {
			assert(false, "open url [%v] driver name [%v] config [%v] error [%v]", strUrl, strDriverName, strConfig, err.Error())
		}
		if err = e.db.Ping(); err != nil {
			assert(false, "ping url [%v] driver name [%v] config [%v] error [%v]", strUrl, strDriverName, strConfig, err.Error())
		}
		e.adapterSqlx = adapterType
	case AdapterCache_Redis:
		var err error
		if e.cache, err = newCache(strDriverName, strConfig); err != nil {
			assert(false, "new cache by driver name [%v] config [%v] error [%v]", strDriverName, strConfig, err.Error())
		}
		e.adapterCache = adapterType
		if len(expireSeconds) > 0 {
			e.expireTime = expireSeconds[0]
		} else {
			e.expireTime = 3600 //one hour expire
		}
	default:
		assert(false, "adapter instance type [%v] url [%s] not support", adapterType, strUrl)
	}

	//log.Struct(e)
	return e
}

func (e *Engine) UseCache() *Engine {
	e.setUseCache(true)
	return e
}

// debug mode on or off
// if debug on, some method will panic if your condition illegal
func (e *Engine) Debug(ok bool) {
	e.setDebug(ok)
}

// orm model
// use to get result set, support single struct object or slice [pointer type]
// notice: will clone a new engine object for orm operations(query/update/insert/upsert)
func (e *Engine) Model(args ...interface{}) *Engine {
	assert(args, "model is nil")
	assert(e.db, "sqlx instance is nil, please call Open method first")

	return e.clone(args...)
}

// set orm query table name
// when your struct type name is not a table name
func (e *Engine) Table(strName string) *Engine {
	assert(strName, "name is nil")
	e.setTableName(strName)
	return e
}

// index which select from cache or update to cache
// if your index is not a primary key, it will create a cache index and pointer to primary key data
// index or data in cache key is 'sqlx:cache:[table name]:[column name]:[column value]', eg. 'sqlx:cache:users:phone:8613055556666'
func (e *Engine) Index(strColumn string, value interface{}) *Engine {

	e.setIndexes(strColumn, value)
	return e
}

// set orm primary key's name, default named 'id'
func (e *Engine) SetPkName(strName string) *Engine {
	assert(strName, "name is nil")
	e.strPkName = strName
	return e
}

func (e *Engine) GetPkName() string {
	return e.strPkName
}

// set orm primary key's value
func (e *Engine) Id(value interface{}) *Engine {

	//TODO @libin sql syntax differences of MySQL/Postgresql/Sqlite/Mssql...
	e.setSelectColumns("*")
	e.setPkValue(value)
	return e
}

// orm select/update columns
func (e *Engine) Select(strColumns ...string) *Engine {
	e.setSelectColumns(strColumns...)
	return e
}

// orm query
// return rows affected and error, if err is not nil must be something wrong
// Model function is must be called before call this function
// notice: use Where function, the records which be updated can not be refreshed to redis/memcached...
func (e *Engine) Where(strWhere string) *Engine {
	assert(strWhere, "string is nil")
	e.setWhere(strWhere)
	return e
}

// set the conflict columns for upsert
// only for postgresql
func (e *Engine) OnConflict(strColumns ...string) *Engine {

	e.setConflictColumns(strColumns...)
	return e
}

// query limit
// Limit(10) - query records limit 10 (mysql/postgres)
func (e *Engine) Limit(args ...int) *Engine {

	//TODO postgresql/mssql limit statement
	nArgs := len(args)
	if nArgs == 1 {
		e.setLimit(fmt.Sprintf("LIMIT %v", args[0]))
	} else if nArgs == 2 {
		e.setLimit(fmt.Sprintf("LIMIT %v,%v", args[0], args[1]))
	}
	return e
}

// query offset (for mysql/postgres)
func (e *Engine) Offset(offset int) *Engine {
	e.setOffset(fmt.Sprintf("OFFSET %v", offset))
	return e
}

// order by [field1,field2...]
func (e *Engine) OrderBy(strColumns ...string) *Engine {
	e.setOrderBy(strColumns...)
	return e
}

// order by [field1,field2...] asc
func (e *Engine) Asc() *Engine {
	e.setAscOrDesc(ORDER_BY_ASC)
	return e
}

// order by [field1,field2...] desc
func (e *Engine) Desc() *Engine {
	e.setAscOrDesc(ORDER_BY_DESC)
	return e
}

// group by [field1,field2...]
func (e *Engine) GroupBy(strColumns ...string) *Engine {
	e.setGroupBy(strColumns...)
	return e
}

// orm query
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) Query() (rowsAffected int64, err error) {
	assert(e.model, "model is nil, please call Model method first")
	e.setOperType(OperType_Query)

	if e.getUseCache() {

		var ok bool
		if rowsAffected, ok = e.queryCache(); ok {
			log.Debugf("query from cache ok, rows affected [%v]", rowsAffected)
			return
		}
	}

	strSqlx := e.makeSqlxString()

	var r *sql.Rows
	if r, err = e.db.Query(strSqlx); err != nil {
		log.Errorf("query [%v] error [%v]", strSqlx, err.Error())
		return
	}

	defer r.Close()
	return e.fetchRows(r)
}

// orm insert
// return last insert id and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) Insert() (lastInsertId int64, err error) {
	assert(e.model, "model is nil, please call Model method first")

	e.setOperType(OperType_Insert)
	var strSqlx string
	strSqlx = e.makeSqlxString()
	var r sql.Result
	r, err = e.db.NamedExec(strSqlx, e.model)
	if err != nil {
		log.Errorf("error %v model %+v", err, e.model)
		return
	}
	lastInsertId, err = r.LastInsertId()
	if err != nil {
		log.Errorf("get last insert id error %v model %+v", err, e.model)
		return
	}
	log.Debugf("lastInsertId = %v", lastInsertId)
	if lastInsertId > 0 {
		e.upsertCache(lastInsertId)
	}
	return
}

// orm insert or update if key(s) conflict
// return last insert id and error, if err is not nil must be something wrong, if your primary key is not a int/int64 type, maybe id return 0
// NOTE: Model function is must be called before call this function and call OnConflict function when you are on postgresql
func (e *Engine) Upsert() (lastInsertId int64, err error) {
	assert(!(e.adapterSqlx == AdapterSqlx_Mssql), "mssql-server un-support insert on duplicate update operation")
	assert(e.model, "model is nil, please call Model method first")
	assert(e.getSelectColumns(), "update columns is not set")

	e.setOperType(OperType_Upsert)
	var strSqlx string
	strSqlx = e.makeSqlxString()

	var r sql.Result
	r, err = e.db.NamedExec(strSqlx, e.model)
	if err != nil {
		log.Errorf("error %v model %+v", err, e.model)
		return
	}
	lastInsertId, err = r.LastInsertId()
	if err != nil {
		log.Errorf("get last insert id error %v model %+v", err, e.model)
		return
	}
	log.Debugf("lastInsertId = %v", lastInsertId)
	if lastInsertId > 0 {
		e.upsertCache(lastInsertId)
	}
	return
}

// orm update from model
// strColumns... if set, columns will be updated, if none all columns in model will be updated except primary key
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) Update() (rowsAffected int64, err error) {
	assert(e.model, "model is nil, please call Model method first")
	assert(e.getSelectColumns(), "update columns is not set, please call Select method")

	e.setOperType(OperType_Update)

	var strSqlx string
	strSqlx = e.makeSqlxString()

	var r sql.Result
	r, err = e.db.Exec(strSqlx)
	if err != nil {
		log.Errorf("error %v model %+v", err, e.model)
		return
	}
	rowsAffected, err = r.RowsAffected()
	if err != nil {
		log.Errorf("get last insert id error [%v] query [%v] model [%+v]", err, strSqlx, e.model)
		return
	}
	log.Debugf("RowsAffected [%v] query [%v]", rowsAffected, strSqlx)

	if rowsAffected > 0 {
		e.updateCache()
	}
	return
}

// use raw sql to query results
// return rows affected and error, if err is not nil must be something wrong
// NOTE: Model function is must be called before call this function
func (e *Engine) QueryRaw(strQuery string, args ...interface{}) (rowsAffected int64, err error) {
	assert(e.db, "sqlx db instance is nil")
	assert(strQuery, "query sql string is nil")
	assert(e.model, "model is nil, please call Model method first")

	e.setOperType(OperType_QueryRaw)

	var r *sqlx.Rows
	if e.isDebug() {
		log.Debugf("query [%v] args %+v", strQuery, args)
	}
	count := strings.Count(strQuery, "?")
	if count > 0 && count == len(args) { //question placeholder exist
		r, err = e.db.Queryx(strQuery, args...)
	} else {
		r, err = e.db.Queryx(fmt.Sprintf(strQuery, args...))
	}

	if err != nil {
		log.Errorf("query [%v] error [%v]", strQuery, err.Error())
		return
	}

	defer r.Close()
	return e.fetchRows(r.Rows)
}

// use raw sql to query results into a map slice (model type is []map[string]string)
// return results and error
// NOTE: Model function is must be called before call this function
func (e *Engine) QueryMap(strQuery string, args ...interface{}) (rowsAffected int64, err error) {
	assert(strQuery, "query sql string is nil")
	assert(e.model, "model is nil, please call Model method first")

	e.setOperType(OperType_QueryMap)
	var r *sqlx.Rows
	if e.isDebug() {
		log.Debugf("query [%v] args %+v", strQuery, args)
	}
	count := strings.Count(strQuery, "?")
	if count > 0 && count == len(args) { //question placeholder exist
		r, err = e.db.Queryx(strQuery, args...)
	} else {
		r, err = e.db.Queryx(fmt.Sprintf(strQuery, args...))
	}

	for r.Next() {
		rowsAffected++
		fetcher, _ := e.getFecther(r.Rows)
		*e.model.(*[]map[string]string) = append(*e.model.(*[]map[string]string), fetcher.mapValues)
	}
	return
}

// use raw sql to insert/update database, results can not be cached to redis/memcached/memory...
// return rows affected and error, if err is not nil must be something wrong
func (e *Engine) ExecRaw(strQuery string, args ...interface{}) (rowsAffected, lastInsertId int64, err error) {
	assert(e.db, "sqlx db instance is nil")
	assert(strQuery, "query sql string is nil")

	e.setOperType(OperType_ExecRaw)

	var r sql.Result
	if e.isDebug() {
		log.Debugf("query [%v] args %+v", strQuery, args)
	}
	count := strings.Count(strQuery, "?")
	if count > 0 && count == len(args) { //question placeholder exist
		r, err = e.db.Exec(strQuery, args...)
	} else {
		r, err = e.db.Exec(fmt.Sprintf(strQuery, args...))
	}
	if err != nil {
		log.Errorf("error [%v] model [%+v]", err, e.model)
		return
	}
	rowsAffected, err = r.RowsAffected()
	if err != nil {
		log.Errorf("get rows affected error [%v] query [%v]", err.Error(), strQuery)
		return
	}
	lastInsertId, err = r.LastInsertId()
	if err != nil {
		log.Errorf("get last insert id error [%v] query [%v]", err.Error(), strQuery)
		return
	}
	return
}

func (e *Engine) Tx(args ...*sqlcaTx) (err error) {

	if err = e.txExec(args...); err != nil {
		log.Errorf("tx exec error [%v]", err.Error())
		return err
	}
	//update to cache
	for _, v := range args {
		e.saveToCache(v.kvs...)
	}
	return nil
}

func (e *Engine) TxRaw(strSQL ...string) (err error) {

	var args []*sqlcaTx
	for _, v := range strSQL {
		args = append(args, newTx(v))
	}

	if err = e.txExec(args...); err != nil {
		log.Errorf("tx exec error [%v]", err.Error())
		return err
	}
	return nil
}

// make orm tx sql: insert
func (e *Engine) ToTxInsert() *sqlcaTx {

	return newTx(e.makeSqlxInsert())
}

// make orm tx sql: upsert
func (e *Engine) ToTxUpsert() *sqlcaTx {

	return newTx(e.makeSqlxUpsert())
}

// make orm tx sql: update
func (e *Engine) ToTxUpdate() *sqlcaTx {

	return newTx(e.makeSqlxUpdate(), e.makeUpdateCache()...)
}

// make orm tx sql: query
func (e *Engine) ToTxQuery() *sqlcaTx {

	return newTx(e.makeSqlxQuery())
}
