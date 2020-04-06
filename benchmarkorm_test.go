//+build benchmarkorm

package appy

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

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
	MAX_IDLE_CONNS               = 32
	MAX_OPEN_CONNS               = 32
	SQL_INSERT_QUERY_PREFIX      = "INSERT INTO users (name, title, fax, web, age, counter) VALUES"
	SQL_INSERT_QUERY_PLACEHOLDER = "(?, ?, ?, ?, ?, ?)"
	SQL_SELECT_QUERY             = "SELECT id, name, title, fax, web, age, counter FROM users WHERE id=?"
	SQL_SELECT_MULTI_QUERY       = "SELECT id, name, title, fax, web, age, counter FROM users WHERE id > 0 LIMIT 100"
	SQL_UPDATE_QUERY             = "UPDATE users SET name=?, title=?, fax=?, web=?, age=?, counter=? WHERE id=?"
)

func newDB() DBer {
	database := "benchmarkorm_appy"
	os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	defer func() {
		os.Unsetenv("DB_URI_PRIMARY")
		os.Unsetenv("DB_MAX_IDLE_CONNS_PRIMARY")
		os.Unsetenv("DB_MAX_OPEN_CONNS_PRIMARY")
	}()

	logger, _, _ := NewFakeLogger()
	dbManager := NewDBManager(logger, &Support{})
	db := dbManager.DB("primary")
	db.ConnectDB("mysql")
	db.DropDB(database)
	db.CreateDB(database)
	db.Connect()
	db.SetMaxIdleConns(MAX_IDLE_CONNS)
	db.SetMaxOpenConns(MAX_OPEN_CONNS)
	db.Exec(SCHEMA)

	return db
}

func newRawDB() *sql.DB {
	database := "benchmarkorm_raw"
	db, err := sql.Open("mysql", "root:whatever@tcp(:13306)/mysql")

	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(MAX_OPEN_CONNS)
	db.SetMaxIdleConns(MAX_IDLE_CONNS)
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
	query := SQL_INSERT_QUERY_PREFIX + strings.Repeat(SQL_INSERT_QUERY_PLACEHOLDER+",", size-1) + SQL_INSERT_QUERY_PLACEHOLDER

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

func rawInsert(db *sql.DB, b *testing.B) (int64, error) {
	result, err := db.Exec(SQL_INSERT_QUERY_PREFIX + "('benchmark', 'just a benchmark', '99991234', 'https://appy.org', 100, 1000)")
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func dbInsert(db DBer, b *testing.B) (int64, error) {
	result, err := db.Exec(SQL_INSERT_QUERY_PREFIX + "('benchmark', 'just a benchmark', '99991234', 'https://appy.org', 100, 1000)")
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func BenchmarkRawInsert(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("%s %s;", SQL_INSERT_QUERY_PREFIX, SQL_INSERT_QUERY_PLACEHOLDER))
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := stmt.Exec("benchmark", "just a benchmark", "99991234", "https://appy.org", 100, 1000)
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

func BenchmarkDBInsert(b *testing.B) {
	db := newDB()
	defer db.Close()

	stmt, err := db.Prepare(fmt.Sprintf("%s %s;", SQL_INSERT_QUERY_PREFIX, SQL_INSERT_QUERY_PLACEHOLDER))
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := stmt.Exec("benchmark", "just a benchmark", "99991234", "https://appy.org", 100, 1000)
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

func BenchmarkRawUpdate(b *testing.B) {
	db := newRawDB()
	defer db.Close()

	id, err := rawInsert(db, b)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	stmt, err := db.Prepare(SQL_UPDATE_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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

	stmt, err := db.Prepare(SQL_UPDATE_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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

	stmt, err := db.Prepare(SQL_SELECT_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	var (
		age, counter          int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		err := stmt.QueryRow(id).Scan(&id, &name, &title, &fax, &web, &age, &counter)
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

	stmt, err := db.Prepare(SQL_SELECT_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	var (
		age, counter          int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
		err := stmt.QueryRow(id).Scan(&id, &name, &title, &fax, &web, &age, &counter)
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

	stmt, err := db.Prepare(SQL_SELECT_MULTI_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	var (
		id, age, counter      int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
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

	stmt, err := db.Prepare(SQL_SELECT_MULTI_QUERY)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
	defer stmt.Close()

	b.ResetTimer()

	var (
		id, age, counter      int
		name, title, fax, web string
	)

	for i := 0; i < b.N; i++ {
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
