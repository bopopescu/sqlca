package sqlca

import (
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/cache"
	"github.com/civet148/gotools/log"
	_ "github.com/go-sql-driver/mysql" //mysql golang driver
	"github.com/jmoiron/sqlx"
)

type Engine struct {
	db               *sqlx.DB          // sqlx instance
	cache            cache.Cache       // beego cache instance
	adapterSqlx      AdapterType       // what's adapter of sqlx
	adapterCache     AdapterType       // what's adapter of cache
	modeType         ModeType          // mode: orm or raw
	operType         OperType          // operate type
	expireTime       int               // cache expire time of seconds
	refreshCache     bool              // can refresh cache ? true=yes false=no
	debug            bool              // debug mode [on/off]
	model            interface{}       // data model [struct object or struct slice]
	dict             map[string]string // data model db dictionary
	strTableName     string            // table name
	strPkName        string            // primary key of table, default 'id'
	strPkValue       string            // primary key's value
	strWhere         string            // where condition to query or update
	strSelectColumns []string          // columns to query: select
	strLimit         string            // limit
	strConflicts     []string          // conflict key on duplicate set (just for postgresql)
}

func NewEngine(debug bool) *Engine {

	return &Engine{
		debug:      debug,
		strPkName:  DEFAULT_PRIMARY_KEY_NAME,
		expireTime: DEFAULT_CAHCE_EXPIRE_SECONDS,
	}
}

// open a sqlx database or cache connection
// @params expireSeconds 	cache data expire seconds, just for AdapterTypeCache_XXX
func (e *Engine) Open(adapterType AdapterType, strDSN string, expireSeconds ...int) *Engine {

	var err error
	switch adapterType {
	case AdapterSqlx_MySQL, AdapterSqlx_Postgres, AdapterSqlx_Sqlite, AdapterSqlx_Mssql:
		strSchema, strDSN := e.getConnUrl(adapterType, strDSN)
		if e.db, err = sqlx.Open(strSchema, strDSN); err != nil {
			e.panic("open schema %v DSN %v original url [%v] error [%v]", strSchema, strDSN, strDSN, err.Error())
		}
		if err = e.db.Ping(); err != nil {
			e.panic("ping database error %v, url %v", err.Error(), strDSN)
		}

		e.adapterSqlx = adapterType
	case AdapterCache_Redis, AdapterCache_Memcached, AdapterCache_Memory, AdapterCache_File:
		// TODO @libin open beego cache conection
		//e.cache = v.(cache.Cache)
		e.adapterCache = adapterType
		if len(expireSeconds) > 0 {
			e.expireTime = expireSeconds[0]
		}
	default:
		assert(false, "open adapter instance type [%v] url [%s] failed", adapterType, strDSN)
	}

	//log.Struct(e)
	return e
}

// attach from a exist sqlx or beego cache instance
// @params expireSeconds 	cache data expire seconds, just for AdapterTypeCache_XXX
func (e *Engine) Attach(adapterType AdapterType, v interface{}, expireSeconds ...int) *Engine {

	switch adapterType {
	case AdapterSqlx_MySQL, AdapterSqlx_Postgres, AdapterSqlx_Sqlite, AdapterSqlx_Mssql:
		{
			assert(!isNilOrFalse(e.db), "already have a [%v] sqlx instance, attach failed", adapterType)
			e.db = v.(*sqlx.DB)
			e.adapterSqlx = adapterType
		}
	case AdapterCache_Redis, AdapterCache_Memcached, AdapterCache_Memory, AdapterCache_File:
		{
			assert(!isNilOrFalse(e.cache), "already have a [%v] beego cache instance, attach failed", adapterType)
			e.cache = v.(cache.Cache)
			e.adapterCache = adapterType
		}
	default:
		assert(false, "adapter type [%v] attach instance failed", adapterType)
		return nil
	}
	return e
}

// get internal sqlx instance
// you can use it to do what you want
func (e *Engine) DB() *sqlx.DB {
	return e.db
}

// get internal cache instance
// you can use it to do what you want
func (e *Engine) Cache() cache.Cache {
	return e.cache
}

// debug mode on or off
// if debug on, some method will panic if your condition illegal
func (e *Engine) Debug(ok bool) {
	e.setDebug(ok)
}

// orm model
// use to get result set, support single struct object or slice [pointer type]
// notice: will clone a new engine object for orm operations
func (e *Engine) Model(v interface{}) *Engine {
	assert(v, "model is nil")
	assert(e.db, "sqlx instance is nil, please call Open or Attach function first")

	return e.clone(v)
}

// set orm query table name
// when your struct type name is not a table name
func (e *Engine) Table(strName string) *Engine {
	assert(strName, "name is nil")
	e.setTableName(strName)
	return e
}

// set orm primary key's name, default named 'id'
func (e *Engine) PkName(strName string) *Engine {
	assert(strName, "name is nil")
	e.setPkName(strName)
	return e
}

// set orm primary key's value
func (e *Engine) Id(value interface{}) *Engine {

	//TODO @libin sql syntax differences of MySQL/Postgresql/Sqlite/Mssql...
	e.setSelectColumns("*")
	e.setPkValue(fmt.Sprintf("%v", value))
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

	e.strConflicts = strColumns
	return e
}

// query limit
// Limit(10) - query records limit 10
// Limit(1, 5) - query records limit 1,5
func (e *Engine) Limit(args ...interface{}) *Engine {

	return e
}

// orm query
// return rows affected and error, if err is not nil must be something wrong
// Model function is must be called before call this function
func (e *Engine) Query() (rows int64, err error) {
	assert(e.model, "model is nil, please call Model function first")

	// TODO @libin Query() implement
	e.operType = OperType_Query
	strSqlx := e.makeSqlxString()

	var r *sql.Rows
	if r, err = e.db.Query(strSqlx); err != nil {
		log.Errorf("query [%v] error [%v]", strSqlx, err.Error())
		return
	}

	defer r.Close()

	for r.Next() {
		var c int64
		if c, err = e.fetchRow(r, e.model); err != nil {
			log.Error("fetchRow error [%v]", err.Error())
			return
		}
		rows += c
	}
	return
}

// orm insert
// return last insert id and error, if err is not nil must be something wrong
// Model function is must be called before call this function
func (e *Engine) Insert() (lastInsertId int64, err error) {
	assert(e.model, "model is nil, please call Model function first")
	// TODO @libin Insert() implement
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
	return
}

// orm insert or update if key(s) conflict
// return last insert id and error, if err is not nil must be something wrong, if your primary key is not a int/int64 type, maybe id return 0
// Model function is must be called before call this function and call OnConflict function when you are on postgresql
func (e *Engine) Upsert(strColumns ...string) (lastInsertId int64, err error) {
	assert(!(e.adapterSqlx == AdapterSqlx_Mssql), "mssql-server un-support insert on duplicate update operation")
	assert(e.model, "model is nil, please call Model function first")
	// TODO @libin Upsert() implement
	e.setOperType(OperType_Upsert)
	e.setSelectColumns(strColumns...)
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
	return
}

// orm update from model
// strColumns... if set, columns will be updated, if none all columns in model will be updated except primary key
// return rows affected and error, if err is not nil must be something wrong
// Model function is must be called before call this function
func (e *Engine) Update(strColumns ...string) (rowsAffected int64, err error) {
	assert(e.model, "model is nil, please call Model function first")
	assert(strColumns, "update columns is not set")
	// TODO @libin Update() implement
	e.setSelectColumns(strColumns...)
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
		log.Errorf("get last insert id error %v model %+v", err, e.model)
		return
	}
	log.Debugf("RowsAffected = %v", rowsAffected)
	return
}

// use raw sql to query results
// return rows and error, if err is not nil must be something wrong
// Model function is must be called before call this function
func (e *Engine) QueryRaw(strQuery string, args ...interface{}) (rowsAffected int64, err error) {
	assert(strQuery, "query sql string is nil")
	assert(e.model, "model is nil, please call Model function first")
	// TODO @libin QuerySQL() implement
	e.setOperType(OperType_QueryRaw)

	var r *sql.Rows
	r, err = e.db.Query(strQuery, args...)
	if err != nil {
		log.Errorf("query [%v] error [%v]", strQuery, err.Error())
		return
	}

	defer r.Close()
	for r.Next() {
		var c int64
		if c, err = e.fetchRow(r, e.model); err != nil {
			log.Errorf("%v", err.Error())
			return
		}
		rowsAffected += c
	}
	log.Debugf("rowsAffected [%v] query [%v]", rowsAffected, strQuery)
	return
}

// use raw sql to insert/update database, results can not be cached to redis/memcached/memory...
// return rows and error, if err is not nil must be something wrong
func (e *Engine) ExecRaw(strQuery string, args ...interface{}) (rowsAffected, lastInsertId int64, err error) {
	assert(strQuery, "query sql string is nil")

	e.setOperType(OperType_ExecRaw)
	var r sql.Result
	r, err = e.db.Exec(strQuery, args...)
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
	log.Debugf("RowsAffected [%v] LastInsertId [%v] query [%v] ", rowsAffected, lastInsertId, strQuery)
	return
}