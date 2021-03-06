package mssql

import (
	"github.com/civet148/gotools/log"
	"github.com/civet148/sqlca"
)

const (
	TABLE_NAME_USERS   = "users"
	TABLE_NAME_CLASSES = "classes"
)

type UserDO struct {
	Id      int32  `db:"id"`
	Name    string `db:"name"`
	Phone   string `db:"phone"`
	Sex     int8   `db:"sex"`
	Email   string `db:"email"`
	Disable int8   `db:"disable"`
	//CreatedAt string `db:"created_at" sqlca:"readonly"`
}

type ClassDo struct {
	Id      int32  `db:"id"`
	UserId  int32  `db:"user_id"`
	ClassNo string `db:"class_no"`
	//CreatedAt string `db:"created_at" sqlca:"readonly"`
}

func Benchmark() {

	e := sqlca.NewEngine("mssql://sa:123456@127.0.0.1:1433/test?instance=SQLEXPRESS&windows=false")
	e.Debug(true) //debug on

	e.Open("redis://127.0.0.1:6379", 3600) //redis alone mode
	//e.Open("redis://123456@127.0.0.1:6379/cluster?db=0&replicate=127.0.0.1:6380,127.0.0.1:6381") //redis cluster mode

	//e.Open("mysql://root:123456@127.0.0.1:3306/test?charset=utf8mb4") //MySQL
	//e.Open("postgres://root:`~!@#$%^&*()-_=+@127.0.0.1:5432/test?sslmode=enable") //postgres
	//e.Open("sqlite:///var/lib/test.db") //sqlite3
	//e.Open("mssql://sa:123456@127.0.0.1:1433/test?instance=SQLEXPRESS&windows=false") //windows MS SQLSERVER

	MSSQL_OrmInsertByModel(e)
	MSSQL_OrmUpsertByModel(e)
	MSSQL_OrmUpdateByModel(e)
	MSSQL_OrmQueryIntoModel(e)
	MSSQL_OrmQueryIntoModelSlice(e)
	MSSQL_OrmUpdateIndexToCache(e)
	MSSQL_OrmSelectMultiTable(e)
	MSSQL_OrmDeleteFromTable(e)
	MSSQL_OrmInCondition(e)
	MSSQL_RawQueryIntoModel(e)
	MSSQL_RawQueryIntoModelSlice(e)
	MSSQL_RawQueryIntoMap(e)
	MSSQL_RawExec(e)
	MSSQL_TxGetExec(e)
	MSSQL_TxRollback(e)
	MSSQL_CustomTag(e)
}

func MSSQL_OrmInsertByModel(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	user := UserDO{
		//Id:    0,
		Name:  "admin",
		Phone: "8618600000000",
		Sex:   1,
		Email: "admin@golang.org",
	}
	if lastInsertId, err := e.Model(&user).Table(TABLE_NAME_USERS).Insert(); err != nil {
		log.Errorf("insert data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("insert data model [%+v] ok, last insert id [%v]", user, lastInsertId)
	}
}

func MSSQL_OrmUpsertByModel(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()
	user := UserDO{
		Id:    0,
		Name:  "lory",
		Phone: "8618677454534",
		Sex:   2,
		Email: "spuset@gmail.com",
	}
	if lastInsertId, err := e.Model(&user).Table(TABLE_NAME_USERS).Select("name", "phone", "email", "sex").Upsert(); err != nil {
		log.Errorf("upsert data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("upsert data model [%+v] ok, last insert id [%v]", user, lastInsertId)
	}
}

func MSSQL_OrmUpdateByModel(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	user := UserDO{
		Id:      1,
		Name:    "john",
		Phone:   "8618699999999",
		Sex:     1,
		Email:   "john@gmail.com",
		Disable: 1,
	}

	//SQL: update users set name='john', phone='8618699999999', sex='1', email='john@gmail.com' where id='1'
	if rowsAffected, err := e.Model(&user).Table(TABLE_NAME_USERS).Select("name", "phone", "email", "sex").Update(); err != nil {
		log.Errorf("update data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("update data model [%+v] ok, rows affected [%v]", user, rowsAffected)
	}
}

func MSSQL_OrmQueryIntoModel(e *sqlca.Engine) {
	log.Enter()
	defer log.Leave()

	user := UserDO{}

	//SQL: select id, name, phone from users where id=1
	//e.Model(&user).Table(TABLE_NAME_USERS).Id(1).Select("id", "name", "phone").Query();

	// select * from users where id=1
	if rowsAffected, err := e.Model(&user).Table(TABLE_NAME_USERS).Id(1).Query(); err != nil {
		log.Errorf("query into data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("query into model [%+v] ok, rows affected [%v]", user, rowsAffected)
	}
}

func MSSQL_OrmQueryIntoModelSlice(e *sqlca.Engine) {
	log.Enter()
	defer log.Leave()

	var users []UserDO

	//SQL: select id, name, phone from users limit 3
	//e.Model(&user).Table(TABLE_NAME_USERS).Select("id", "name", "phone").Limit(3).Query();

	//SQL: select * from users limit 3
	if rowsAffected, err := e.Model(&users).Table(TABLE_NAME_USERS).Limit(3).Query(); err != nil {
		log.Errorf("query into data model [%+v] error [%v]", users, err.Error())
	} else {

		if len(users) == 0 {
			log.Errorf("query into model failed, rows affected [%v]", rowsAffected)
		} else {
			log.Debugf("query into model [%+v] ok, rows affected [%v]", users, rowsAffected)
		}
	}
}

func MSSQL_RawQueryIntoModel(e *sqlca.Engine) {
	log.Enter()
	defer log.Leave()

	user := UserDO{}

	//SQL: select * from users where id=1
	if rowsAffected, err := e.Model(&user).QueryRaw("select * from users where id=?", 1); err != nil {
		log.Errorf("query into data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("query into model [%+v] ok, rows affected [%v]", user, rowsAffected)
	}
}

func MSSQL_RawQueryIntoModelSlice(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	var users []UserDO

	//SQL: select * from users where id < 5
	if rowsAffected, err := e.Model(&users).QueryRaw("select * from %v where id < %v", TABLE_NAME_USERS, 5); err != nil {
		log.Errorf("query into data model [%+v] error [%v]", users, err.Error())
	} else {
		log.Debugf("query into model [%+v] ok, rows affected [%v]", users, rowsAffected)
	}
}

func MSSQL_RawQueryIntoMap(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	var users []map[string]string

	//SQL: select * from users where id < 5
	if rowsAffected, err := e.Model(&users).QueryMap("select * from %v where id < %v", TABLE_NAME_USERS, 5); err != nil {
		log.Errorf("query into map [%+v] error [%v]", users, err.Error())
	} else {
		log.Debugf("query into map [%+v] ok, rows affected [%v]", users, rowsAffected)
	}
}

func MSSQL_RawExec(e *sqlca.Engine) {

	//e.ExecRaw("UPDATE %v SET name='duck' WHERE id='%v'", TABLE_NAME_USERS, 2) //it will work well as question placeholder
	rowsAffected, lasteInsertId, err := e.ExecRaw("UPDATE users SET name=? WHERE id=?", "duck", 1)
	if err != nil {
		log.Errorf("exec raw sql error [%v]", err.Error())
	} else {
		log.Debugf("exec raw sql ok, rows affected [%v] last insert id [%v]", rowsAffected, lasteInsertId)
	}
}

func MSSQL_OrmUpdateIndexToCache(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	user := UserDO{
		Id:    1,
		Name:  "john3",
		Phone: "8615011111114",
		Sex:   1,
		Email: "john3@gmail.com",
	}

	//SQL: update users set name='john3', phone='8615011111114', sex='1', email='john3@gmail.com' where id='1'
	//index: name, phone
	//redis key:  sqlx:cache:[table]:[column]:[column value]
	if rowsAffected, err := e.Model(&user).
		Table(TABLE_NAME_USERS).
		Distinct().
		Select("name", "phone", "email", "sex").
		Cache("name", "phone").
		Update(); err != nil {
		log.Errorf("update data model [%+v] error [%v]", user, err.Error())
	} else {
		log.Debugf("update data model [%+v] ok, rows affected [%v]", user, rowsAffected)
	}
}

func MSSQL_OrmSelectMultiTable(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	type UserClass struct {
		UserId   int32  `db:"user_id"`
		UserName string `db:"user_name"`
		Phone    string `db:"phone"`
		ClassNo  string `db:"class_no"`
	}
	var ucs []UserClass
	//SQL: SELECT a.*, b.class_no FROM users a, classes b WHERE a.id=b.user_id AND a.id=3
	_, err := e.Model(&ucs).
		Distinct().
		Select("a.id as user_id", "a.name", "a.phone", "b.class_no").
		Table("users a", "classes b").
		Where("a.id=b.user_id").
		And("a.id=?", 3).
		Query()
	if err != nil {
		log.Errorf("query error [%v]", err.Error())
	} else {
		log.Debugf("user class info [%+v]", ucs)
	}
}

func MSSQL_OrmDeleteFromTable(e *sqlca.Engine) {

	log.Enter()
	defer log.Leave()

	user := UserDO{
		Id: 1000,
	}
	//delete from data model
	if rows, err := e.Model(&user).Table(TABLE_NAME_USERS).Delete(); err != nil {
		log.Errorf("delete from table error [%v]", err.Error())
	} else {
		log.Debugf("delete from table ok, affected rows [%v]", rows)
	}

	//delete from where condition (without data model)
	if rows, err := e.Table(TABLE_NAME_USERS).Where("id > 1001").Delete(); err != nil {
		log.Errorf("delete from table error [%v]", err.Error())
	} else {
		log.Debugf("delete from table ok, affected rows [%v]", rows)
	}

	//delete from primary key 'id' and value (without data model)
	if rows, err := e.Table(TABLE_NAME_USERS).Id(1002).Where("disable=1").Delete(); err != nil {
		log.Errorf("delete from table error [%v]", err.Error())
	} else {
		log.Debugf("delete from table ok, affected rows [%v]", rows)
	}
}

func MSSQL_OrmInCondition(e *sqlca.Engine) {
	log.Enter()
	defer log.Leave()

	var users []UserDO
	//SQL: select * from users where id > 2 and id in (1,3,6,7) and disable in (0,1)
	if rows, err := e.Model(&users).
		Table(TABLE_NAME_USERS).
		Select("*").
		Where("id > 2").
		In("id", 1, 3, 6, 7).
		In("disable", 0, 1).
		Query(); err != nil {
		log.Errorf("select from table by in condition error [%v]", err.Error())
	} else {
		log.Debugf("select from table by in condition ok, affected rows [%v]", rows)
	}
}

func MSSQL_TxGetExec(e *sqlca.Engine) (err error) {
	log.Enter()
	defer log.Leave()

	var tx *sqlca.Engine
	//transaction: select user id form users where phone is '8618600000000' and update users disable to 1 by user id
	if tx, err = e.TxBegin(); err != nil {
		log.Errorf("TxBegin error [%v]", err.Error())
		return
	}

	var UserId int32

	//query results into base variants
	_, err = tx.TxGet(&UserId, "SELECT id FROM users WHERE phone='%v'", "8618600000000")
	if err != nil {
		log.Errorf("TxGet error %v", err.Error())
		_ = tx.TxRollback()
		return
	}
	var lastInsertId, rowsAffected int64
	if UserId == 0 {
		log.Warnf("select id users by phone number but user not exist")
		_ = tx.TxRollback()
		return
	}
	log.Debugf("base variant of user id [%+v]", UserId)
	lastInsertId, rowsAffected, err = tx.TxExec("UPDATE users SET disable=? WHERE id=?", 1, UserId)
	if err != nil {
		log.Errorf("TxExec error %v", err.Error())
		_ = tx.TxRollback()
		return
	}
	log.Debugf("user id [%v] disabled, last insert id [%v] rows affected [%v]", UserId, lastInsertId, rowsAffected)

	//query results into a struct object or slice
	var dos []UserDO
	_, err = tx.TxGet(&dos, "SELECT TOP 5 * FROM users WHERE disable=1")
	if err != nil {
		log.Errorf("TxGet error %v", err.Error())
		_ = tx.TxRollback()
		return
	}
	for _, do := range dos {
		log.Debugf("struct user data object [%+v]", do)
	}

	if err = tx.TxCommit(); err != nil {
		log.Errorf("TxCommit error [%v]", err.Error())
		return
	}
	return
}

func MSSQL_TxRollback(e *sqlca.Engine) (err error) {

	log.Enter()
	defer log.Leave()

	var tx *sqlca.Engine
	//transaction: insert and rollback
	if tx, err = e.TxBegin(); err != nil {
		log.Errorf("TxBegin error [%v]", err.Error())
		return
	}

	_, _, err = tx.TxExec("INSERT INTO users(id, name, phone, sex, email) VALUES(1, 'john3', '8618600000000', 2, 'john3@gmail.com')")
	if err != nil {
		log.Errorf("TxExec error %v, rollback", err.Error())
		_ = tx.TxRollback()
		return
	}

	if err = tx.TxCommit(); err != nil {
		log.Errorf("TxCommit error [%v]", err.Error())
		return
	}
	return
}

func MSSQL_CustomTag(e *sqlca.Engine) {
	type CustomUser struct {
		Id    int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"` // protobuf tag
		Name  string `json:"name"`                                                // json tag
		Phone string `db:"phone"`                                                 // db tag
	}

	var users []CustomUser
	//add custom tag
	e.SetCustomTag(sqlca.TAG_NAME_PROTOBUF, sqlca.TAG_NAME_JSON)
	if count, err := e.Model(&users).
		Table(TABLE_NAME_USERS).
		Where("id < ?", 5).
		Query(); err != nil {
		log.Errorf("custom tag query error [%v]", err.Error())
	} else {
		log.Debugf("custom tag query results %+v rows [%v]", users, count)
	}
}
