package record

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/appist/appy/support"

	// Automatically import mysql driver to make it easier for appy's users.
	_ "github.com/go-sql-driver/mysql"
)

const (
	SCHEMA = `
CREATE TABLE users (
	id int(11) NOT NULL AUTO_INCREMENT,
	name varchar(255) NOT NULL,
	title varchar(255) NOT NULL,
	fax varchar(255) NOT NULL,
	web varchar(255) NOT NULL,
	age int(11) NOT NULL,
	counter bigint(20) NOT NULL,
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
`
	MaxIdleConns              = 32
	MaxOpenConns              = 32
	SQLInsertQueryPrefix      = "INSERT INTO users (name, title, fax, web, age, counter) VALUES"
	SQLInsertQueryPlaceholder = "(?, ?, ?, ?, ?, ?)"
	SQLSelectQuery            = "SELECT id, name, title, fax, web, age, counter FROM users WHERE id=?"
	SQLSelectMultiQuery       = "SELECT id, name, title, fax, web, age, counter FROM users WHERE id > 0 LIMIT 100"
	SQLUpdateQuery            = "UPDATE users SET name=?, title=?, fax=?, web=?, age=?, counter=? WHERE id=?"
)

type BenchmarkUser struct {
	Modeler `masters:"primary" replicas:"" tableName:"users" primaryKeys:"id"`
	ID      int64 `db:"id" orm:"auto_increment:true"`
	Age     int
	Fax     string
	Name    string
	Title   string
	Web     string
	Counter int64
}

func newDB() DBer {
	database := "benchmarkrecord_db"
	os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	os.Setenv("DB_MAX_IDLE_CONNS_PRIMARY", strconv.Itoa(MaxIdleConns))
	os.Setenv("DB_MAX_OPEN_CONNS_PRIMARY", strconv.Itoa(MaxOpenConns))
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
	}()

	logger, _, _ := support.NewTestLogger()
	dbManager := NewEngine(logger)
	db := dbManager.DB("primary")
	db.DropDB(database)
	db.CreateDB(database)
	db.Connect()
	db.Exec(SCHEMA)

	return db
}

func newOrmDBManager() *Engine {
	database := "benchmarkrecord_orm"
	os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	os.Setenv("DB_MAX_IDLE_CONNS_PRIMARY", strconv.Itoa(MaxIdleConns))
	os.Setenv("DB_MAX_OPEN_CONNS_PRIMARY", strconv.Itoa(MaxOpenConns))
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
	}()

	logger, _, _ := support.NewTestLogger()
	dbManager := NewEngine(logger)
	db := dbManager.DB("primary")
	db.DropDB(database)
	db.CreateDB(database)
	db.Connect()
	db.Exec(SCHEMA)

	return dbManager
}

func newRawDB() *sql.DB {
	database := "benchmarkrecord_raw"
	db, err := sql.Open("mysql", "root:whatever@tcp(:13306)/mysql")

	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxIdleConns(MaxIdleConns)
	db.SetMaxOpenConns(MaxOpenConns)
	db.Exec(fmt.Sprintf("DROP DATABASE %s;", database))
	db.Exec(fmt.Sprintf("CREATE DATABASE %s;", database))
	db.Exec(fmt.Sprintf("USE %s;", database))
	_, err = db.Exec(SCHEMA)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func newQueryWithArgs() (string, []interface{}) {
	size := 100
	query := SQLInsertQueryPrefix + strings.Repeat(SQLInsertQueryPlaceholder+",", size-1) + SQLInsertQueryPlaceholder

	args := []interface{}{}
	for i := 0; i < size; i++ {
		args = append(args, "benchmark")
		args = append(args, "just a benchmark")
		args = append(args, "99991234")
		args = append(args, "https://appy.org")
		args = append(args, 100)
		args = append(args, 1000)
	}

	return query, args
}

func dbInsert(db DBer, b *testing.B) (int64, error) {
	result, err := db.Exec(SQLInsertQueryPrefix + "('benchmark', 'just a benchmark', '99991234', 'https://appy.org', 100, 1000)")
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func ormInsert(dbManager *Engine, b *testing.B) (int64, error) {
	user := BenchmarkUser{
		Name:    "benchmark",
		Title:   "just a benchmark",
		Fax:     "99991234",
		Web:     "https://appy.org",
		Age:     100,
		Counter: 1000,
	}
	model := NewModel(dbManager, &user)

	_, err := model.Create().Exec(nil, false)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func rawInsert(db *sql.DB, b *testing.B) (int64, error) {
	result, err := db.Exec(SQLInsertQueryPrefix + "('benchmark', 'just a benchmark', '99991234', 'https://appy.org', 100, 1000)")
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func BenchmarkRawInsert(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := rawInsert(db, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkDBInsert(b *testing.B) {
	db := newDB()
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := dbInsert(db, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkOrmInsert(b *testing.B) {
	dbManager := newOrmDBManager()
	defer dbManager.DB("primary").Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ormInsert(dbManager, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkRawInsertMulti(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	query, args := newQueryWithArgs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := db.Exec(query, args...)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		_, err = result.LastInsertId()
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkDBInsertMulti(b *testing.B) {
	db := newDB()
	defer db.Close()

	query, args := newQueryWithArgs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := db.Exec(query, args...)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		_, err = result.LastInsertId()
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkOrmInsertMulti(b *testing.B) {
	dbManager := newOrmDBManager()
	defer dbManager.DB("primary").Close()

	users := []BenchmarkUser{}
	for i := 0; i < b.N; i++ {
		users = append(users, BenchmarkUser{
			Name:    "benchmark",
			Title:   "just a benchmark",
			Fax:     "99991234",
			Web:     "https://appy.org",
			Age:     100,
			Counter: 1000,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		model := NewModel(dbManager, &users)
		_, err := model.Create().Exec(nil, false)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkRawUpdate(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	id, err := rawInsert(db, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLUpdateQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		_, err = stmt.Exec("benchmark", "just a benchmark", "99991234", "https://appy.org", 100, 1000, id)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkDBUpdate(b *testing.B) {
	db := newDB()
	defer db.Close()

	id, err := dbInsert(db, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLUpdateQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		_, err = stmt.Exec("benchmark", "just a benchmark", "99991234", "https://appy.org", 100, 1000, id)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkRawRead(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	id, err := rawInsert(db, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	b.ResetTimer()

	var (
		age, counter          int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLSelectQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		err = stmt.QueryRow(id).Scan(&id, &name, &title, &fax, &web, &age, &counter)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkDBRead(b *testing.B) {
	db := newDB()
	defer db.Close()

	id, err := dbInsert(db, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	b.ResetTimer()

	var (
		age, counter          int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLSelectQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		err = stmt.QueryRow(id).Scan(&id, &name, &title, &fax, &web, &age, &counter)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkOrmRead(b *testing.B) {
	dbManager := newOrmDBManager()
	defer dbManager.DB("primary").Close()

	id, err := ormInsert(dbManager, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	b.ResetTimer()

	var user BenchmarkUser
	for i := 0; i < b.N; i++ {
		model := NewModel(dbManager, &user)
		count, err := model.Where("id = ?", id).Find().Exec(nil, false)
		if count != 1 {
			fmt.Println(errors.New("count should equal to 1"))
			b.FailNow()
		}

		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkRawReadSlice(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	for i := 0; i < 100; i++ {
		_, err := rawInsert(db, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}

	b.ResetTimer()

	var (
		id, age, counter      int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLSelectMultiQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		for j := 0; rows.Next() && j < 100; j++ {
			err = rows.Scan(&id, &name, &title, &fax, &web, &age, &counter)
			if err != nil {
				fmt.Println(err)
				b.FailNow()
			}
		}

		if err = rows.Err(); err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		if err = rows.Close(); err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkDBReadSlice(b *testing.B) {
	db := newDB()
	defer db.Close()

	for i := 0; i < 100; i++ {
		_, err := dbInsert(db, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}

	b.ResetTimer()

	var (
		id, age, counter      int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare(SQLSelectMultiQuery)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		for j := 0; rows.Next() && j < 100; j++ {
			err = rows.Scan(&id, &name, &title, &fax, &web, &age, &counter)
			if err != nil {
				fmt.Println(err)
				b.FailNow()
			}
		}

		if err = rows.Err(); err != nil {
			fmt.Println(err)
			b.FailNow()
		}

		if err = rows.Close(); err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}

func BenchmarkOrmReadSlice(b *testing.B) {
	dbManager := newOrmDBManager()
	defer dbManager.DB("primary").Close()

	for i := 0; i < 100; i++ {
		_, err := ormInsert(dbManager, b)
		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}

	b.ResetTimer()

	var user BenchmarkUser
	for i := 0; i < b.N; i++ {
		model := NewModel(dbManager, &user)
		count, err := model.Where("id > ?", 0).Limit(100).Find().Exec(nil, false)
		if count != 1 {
			fmt.Println(errors.New("count should equal to 1"))
			b.FailNow()
		}

		if err != nil {
			fmt.Println(err)
			b.FailNow()
		}
	}
}
