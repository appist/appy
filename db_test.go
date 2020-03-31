package appy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lib/pq"
)

type DBSuite struct {
	TestSuite
	buffer          *bytes.Buffer
	dbManager       *DBManager
	writer          *bufio.Writer
	logger          *Logger
	support         Supporter
	mysqlDB, psqlDB DBer
}

func (s *DBSuite) SetupTest() {
	s.logger, s.buffer, s.writer = NewFakeLogger()
	s.support = &Support{}
}

func (s *DBSuite) setupMySQL(database string) {
	os.Setenv("DB_URI_MYSQL", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
	defer os.Unsetenv("DB_URI_MYSQL")

	s.dbManager = NewDBManager(s.logger, s.support)
	s.mysqlDB = s.dbManager.DB("mysql")
	s.Equal("0.0.0.0", s.mysqlDB.Config().Host)
	s.Equal("13306", s.mysqlDB.Config().Port)
	s.Equal(database, s.mysqlDB.Config().Database)

	s.mysqlDB.ConnectDB("mysql")
	s.mysqlDB.DropDB(database)
	s.mysqlDB.CreateDB(database)
	s.mysqlDB.Connect()
	s.mysqlDB.Exec(`
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`)
}

func (s *DBSuite) setupPostgreSQL(database string) {
	os.Setenv("DB_URI_POSTGRES", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
	defer os.Unsetenv("DB_URI_POSTGRES")

	s.dbManager = NewDBManager(s.logger, s.support)
	s.psqlDB = s.dbManager.DB("postgres")
	s.Equal("0.0.0.0", s.psqlDB.Config().Host)
	s.Equal("15432", s.psqlDB.Config().Port)
	s.Equal(database, s.psqlDB.Config().Database)

	s.psqlDB.ConnectDB("postgres")
	s.psqlDB.DropDB(database)
	s.psqlDB.CreateDB(database)
	s.psqlDB.Connect()
	s.psqlDB.Exec(`
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
);
`)
}

func (s *DBSuite) TestConnect() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:@0.0.0.0:15432/test_connect?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)
	err := db.Connect()
	s.Equal("password authentication failed for user \"postgres\"", err.(*pq.Error).Message)
}

func (s *DBSuite) TestConnectDB() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/test_connect?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)
	db.config.URI = "...postgres://"
	err := db.ConnectDB("foobar")
	s.Equal(err.Error(), `parse "...postgres://": first path segment in URL cannot contain colon`)

	db.config.URI = "postgres://postgres:whatever@0.0.0.0:15432/test_connect?sslmode=disable&connect_timeout=5"
	err = db.ConnectDB("foobar")
	s.Equal(`pq: database "foobar" does not exist`, err.(*pq.Error).Error())
}

func (s *DBSuite) TestCreateDB() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)
	err := db.Connect()
	s.Nil(err)

	err = db.CreateDB("postgres")
	s.Equal(`pq: database "postgres" already exists`, err.Error())
}

func (s *DBSuite) TestDropDB() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)
	err := db.Connect()
	s.Nil(err)

	err = db.DropDB("foobar")
	s.Equal(`pq: database "foobar" does not exist`, err.Error())
}

func (s *DBSuite) TestGenerateMigration() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	oldDBMigratePath := dbMigratePath
	dbMigratePath = "tmp/" + dbMigratePath
	defer func() { dbMigratePath = oldDBMigratePath }()

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)

	// Test gen:migration without DBTx
	err := db.GenerateMigration("CreateUsers", "primary", false)
	s.Nil(err)

	files := []string{}
	_ = filepath.Walk(dbMigratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err := ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigration(")

	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)

	// Test gen:migration with DBTx
	err = db.GenerateMigration("CreateUsers", "primary", true)
	s.Nil(err)

	files = []string{}
	_ = filepath.Walk(dbMigratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err = ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigrationTx(")

	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)
}

func (s *DBSuite) TestTransactionForMySQL() {
	database := "test_transaction_for_mysql"
	s.setupMySQL(database)

	tx, err := s.mysqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES(?);", "foobar")
	s.Nil(err)
	tx.Rollback()

	var count int
	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)

	tx, err = s.mysqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES(?);", "foobar")
	s.Nil(err)
	tx.Commit()

	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)
}

func (s *DBSuite) TestTransactionForPostgreSQL() {
	database := "test_transaction_for_postgresql"
	s.setupPostgreSQL(database)

	tx, err := s.psqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES($1);", "foobar")
	s.Nil(err)
	tx.Rollback()

	var count int
	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)

	tx, err = s.psqlDB.Begin()
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES($1);", "foobar")
	s.Nil(err)
	tx.Commit()

	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)
}

func (s *DBSuite) TestTransactionWithContextForMySQL() {
	database := "test_transaction_with_context_for_mysql"
	s.setupMySQL(database)

	ctx, cancel := context.WithCancel(context.Background())
	tx, err := s.mysqlDB.BeginContext(ctx, nil)
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES(?);", "barfoo")
	cancel()
	s.Nil(err)
	s.NotNil(tx.Commit())

	var count int
	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)
}

func (s *DBSuite) TestTransactionWithContextForPostgreSQL() {
	database := "test_transaction_with_context_for_postgresql"
	s.setupPostgreSQL(database)

	ctx, cancel := context.WithCancel(context.Background())
	tx, err := s.psqlDB.BeginContext(ctx, nil)
	s.Nil(err)

	_, err = tx.Exec("INSERT INTO users VALUES($1);", "barfoo")
	cancel()
	s.Nil(err)
	s.NotNil(tx.Commit())

	var count int
	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)
}

func (s *DBSuite) TestMigrateSeedRollbackForMySQL() {
	database := "test_migrate_seed_rollback_for_mysql"
	s.setupMySQL(database)

	oldDBMigratePath := dbMigratePath
	dbMigratePath = "tmp/" + dbMigratePath
	defer func() { dbMigratePath = oldDBMigratePath }()

	err := s.mysqlDB.Connect()
	s.Nil(err)

	err = s.mysqlDB.RegisterMigrationTx(
		func(tx *DBTx) error {
			_, err := tx.Exec(``)
			return err
		},
		func(tx *DBTx) error {
			_, err := tx.Exec(``)
			return err
		},
		"create_orders",
	)
	s.Equal("invalid filename '\"create_orders\"', a valid example: 20060102150405_create_users.go", err.Error())

	err = s.mysqlDB.RegisterMigrationTx(
		func(tx *DBTx) error {
			_, err := tx.Exec(`CREATE orders`)

			return err
		},
		func(tx *DBTx) error {
			_, err := tx.Exec(`DROP TABLE IF EXISTS orders;`)

			return err
		},
		"20200201165238_create_orders",
	)
	s.Nil(err)

	err = s.mysqlDB.Migrate()
	s.Equal("Error 1064: You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'orders' at line 1", err.Error())

	s.mysqlDB.(*DB).migrations = []*DBMigration{}
	err = s.mysqlDB.RegisterMigrationTx(
		func(tx *DBTx) error {
			_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS orders (
					id int NOT NULL,
					username VARCHAR(255) NOT NULL
				);
			`)

			return err
		},
		func(tx *DBTx) error {
			_, err := tx.Exec(`DROP TABLE IF EXISTS orders;`)

			return err
		},
		"20200201165238_create_orders",
	)
	s.Nil(err)

	err = s.mysqlDB.RegisterMigration(
		func(db *DB) error {
			_, err := db.Exec(`CREATE INDEX orders_on_username ON orders(username);`)
			return err
		},
		func(db *DB) error {
			_, err := db.Exec(`ALTER TABLE orders DROP INDEX orders_on_username;`)
			return err
		},
		"20200202165238_add_orders_on_username_index",
	)
	s.Nil(err)

	err = s.mysqlDB.Migrate()
	s.Nil(err)

	migrations, err := s.mysqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("up", migrations[1][0])

	// Test db:schema:dump after db:migrate
	err = s.mysqlDB.DumpSchema(database)
	s.Nil(err)

	schemaPath := dbMigratePath + "/" + database + "/schema.go"
	_, err = os.Stat(schemaPath)
	s.Equal(false, os.IsNotExist(err))

	buf, err := ioutil.ReadFile(schemaPath)
	s.Nil(err)
	s.Contains(string(buf), fmt.Sprintf("package %s", database))
	s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))
	s.Contains(string(buf), "CREATE TABLE orders (")
	s.Contains(string(buf), "KEY orders_on_username (username)")
	s.Contains(string(buf), "('20200201165238'),")
	s.Contains(string(buf), "('20200202165238');")

	// Test db:seed
	s.mysqlDB.RegisterSeedTx(
		func(tx *DBTx) error {
			usernames := []string{"john doe", "james bond"}

			for idx, username := range usernames {
				_, err := tx.Exec(`INSERT INTO orders VALUES(?, ?);`, idx, username)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	err = s.mysqlDB.Seed()
	s.Nil(err)

	var count int
	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM orders")
	s.Nil(err)
	s.Equal(2, count)

	// Test db:rollback
	err = s.mysqlDB.Rollback()
	s.Nil(err)

	migrations, err = s.mysqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("down", migrations[1][0])

	err = s.mysqlDB.Rollback()
	s.Nil(err)

	migrations, err = s.mysqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("down", migrations[0][0])
	s.Equal("down", migrations[1][0])

	// Test db:schema:dump after db:rollback
	err = s.mysqlDB.DumpSchema(database)
	s.Nil(err)

	buf, err = ioutil.ReadFile(schemaPath)
	s.Nil(err)
	s.Contains(string(buf), fmt.Sprintf("package %s", database))
	s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))
	s.NotContains(string(buf), "CREATE TABLE orders (")
	s.NotContains(string(buf), "KEY orders_on_username (username)")
	s.NotContains(string(buf), "('20200201165238'),")
	s.NotContains(string(buf), "('20200202165238');")

	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)
}

func (s *DBSuite) TestMigrateSeedRollbackForPostgreSQL() {
	database := "test_migrate_seed_rollback_for_postgresql"
	s.setupPostgreSQL(database)

	oldDBMigratePath := dbMigratePath
	dbMigratePath = "tmp/" + dbMigratePath
	defer func() { dbMigratePath = oldDBMigratePath }()

	err := s.psqlDB.Connect()
	s.Nil(err)

	s.psqlDB.RegisterMigrationTx(
		func(tx *DBTx) error {
			_, err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS orders (
						id int NOT NULL,
						username VARCHAR(255) NOT NULL
					);
				`)

			return err
		},
		func(tx *DBTx) error {
			_, err := tx.Exec(`DROP TABLE IF EXISTS orders;`)

			return err
		},
		"20200201165238_create_orders",
	)

	s.psqlDB.RegisterMigration(
		func(db *DB) error {
			_, err := db.Exec(`CREATE INDEX CONCURRENTLY orders_on_username ON orders(username);`)
			return err
		},
		func(db *DB) error {
			_, err := db.Exec(`DROP INDEX orders_on_username;`)
			return err
		},
		"20200202165238_add_orders_on_username_index",
	)

	err = s.psqlDB.Migrate()
	s.Nil(err)

	migrations, err := s.psqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("up", migrations[1][0])

	// Test db:schema:dump after db:migrate
	err = s.psqlDB.DumpSchema(database)
	s.Nil(err)

	schemaPath := dbMigratePath + "/" + database + "/schema.go"
	_, err = os.Stat(schemaPath)
	s.Equal(false, os.IsNotExist(err))

	buf, err := ioutil.ReadFile(schemaPath)
	s.Nil(err)
	s.Contains(string(buf), fmt.Sprintf("package %s", database))
	s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))
	s.Contains(string(buf), "CREATE TABLE IF NOT EXISTS public.orders (")
	s.Contains(string(buf), "CREATE INDEX orders_on_username ON public.orders USING btree (username)")
	s.Contains(string(buf), "('20200201165238'),")
	s.Contains(string(buf), "('20200202165238');")

	// Test db:seed
	s.psqlDB.RegisterSeedTx(
		func(tx *DBTx) error {
			usernames := []string{"john doe", "james bond"}

			for idx, username := range usernames {
				_, err := tx.Exec(`INSERT INTO orders VALUES($1, $2);`, idx, username)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	err = s.psqlDB.Seed()
	s.Nil(err)

	var count int
	err = s.psqlDB.Get(&count, "SELECT COUNT(*) FROM orders")
	s.Nil(err)
	s.Equal(2, count)

	// Test db:rollback
	err = s.psqlDB.Rollback()
	s.Nil(err)

	migrations, err = s.psqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("up", migrations[0][0])
	s.Equal("down", migrations[1][0])

	err = s.psqlDB.Rollback()
	s.Nil(err)

	migrations, err = s.psqlDB.MigrateStatus()
	s.Nil(err)
	s.Equal(2, len(migrations))
	s.Equal("down", migrations[0][0])
	s.Equal("down", migrations[1][0])

	// Test db:schema:dump after db:rollback
	err = s.psqlDB.DumpSchema(database)
	s.Nil(err)

	buf, err = ioutil.ReadFile(schemaPath)
	s.Nil(err)
	s.Contains(string(buf), fmt.Sprintf("package %s", database))
	s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))
	s.NotContains(string(buf), "CREATE TABLE IF NOT EXISTS public.orders (")
	s.NotContains(string(buf), "CREATE INDEX orders_on_username ON public.orders USING btree (username)")
	s.NotContains(string(buf), "('20200201165238'),")
	s.NotContains(string(buf), "('20200202165238');")

	err = os.RemoveAll(dbMigratePath)
	s.Nil(err)
}

func (s *DBSuite) TestOperations() {
	database := "test_db_operations"
	s.setupMySQL(database)

	err := s.mysqlDB.Connect()
	s.Nil(err)

	_, err = s.mysqlDB.Exec("SELECT * FROM users;")
	s.Nil(err)

	var count int
	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(0, count)

	_, err = s.mysqlDB.NamedExec(`INSERT INTO users (username) VALUES (:username);`,
		map[string]interface{}{
			"username": "Smith",
		},
	)
	s.Nil(err)

	err = s.mysqlDB.Get(&count, "SELECT COUNT(*) FROM users;")
	s.Nil(err)
	s.Equal(1, count)

	_, err = s.mysqlDB.NamedQuery(`SELECT * FROM users WHERE username=:username;`,
		map[string]interface{}{
			"username": "Smith",
		},
	)
	s.Nil(err)

	type User struct {
		Username string `db:"username"`
	}

	user := User{Username: "Smith"}
	users := []User{}
	namedStmt, err := s.mysqlDB.PrepareNamed(`SELECT * FROM users WHERE username = :username;`)
	s.Nil(err)

	err = namedStmt.Select(&users, user)
	s.Nil(err)
	s.Equal(1, len(users))

	row := s.mysqlDB.QueryRow(`SELECT * FROM users WHERE username=?;`, "Smith")
	var username string
	err = row.Scan(&username)
	s.Nil(err)
	s.Equal("Smith", username)

	stmt, err := s.mysqlDB.Prepare(`SELECT * FROM users WHERE username=?;`)
	row = stmt.QueryRow("Smith")
	s.Nil(err)
	row.Scan(&user)
	s.Equal("Smith", user.Username)

	var usernames []string
	err = s.mysqlDB.Select(&usernames, "SELECT username FROM users;")
	s.Nil(err)
	s.Equal(1, len(usernames))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = s.mysqlDB.ExecContext(ctx, "SELECT * FROM users;")
	s.Equal("context canceled", err.Error())

	err = s.mysqlDB.GetContext(ctx, &count, "SELECT COUNT(*) FROM users;")
	s.Equal("context canceled", err.Error())

	_, err = s.mysqlDB.NamedExecContext(ctx, `INSERT INTO users (username) VALUES (:username);`,
		map[string]interface{}{
			"username": "Smith",
		},
	)
	s.Equal("context canceled", err.Error())

	_, err = s.mysqlDB.NamedQueryContext(ctx, `SELECT * FROM users WHERE username=:username;`,
		map[string]interface{}{
			"username": "Smith",
		},
	)
	s.Equal("context canceled", err.Error())

	users = []User{}
	_, err = s.mysqlDB.PrepareNamedContext(ctx, `SELECT * FROM users WHERE username = :username;`)
	s.Equal("context canceled", err.Error())

	_, err = s.mysqlDB.Query(`SELECT * FROM users;`)
	s.Nil(err)

	_, err = s.mysqlDB.QueryContext(ctx, `SELECT * FROM users;`)
	s.Equal("context canceled", err.Error())

	usernames = []string{}
	err = s.mysqlDB.SelectContext(ctx, &usernames, "SELECT username FROM users;")
	s.Equal("context canceled", err.Error())

	row = s.mysqlDB.QueryRowContext(ctx, `SELECT * FROM users WHERE username=?;`, "Smith")
	err = row.Scan(&username)
	s.Equal("context canceled", err.Error())

	_, err = s.mysqlDB.PrepareContext(ctx, `SELECT * FROM users WHERE username=?;`)
	s.Equal("context canceled", err.Error())
}

func (s *DBSuite) TestSetSchema() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	dbConfigs, errs := parseDBConfig(s.support)
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger, s.support)
	s.Equal("", db.Schema())

	db.SetSchema(`foobar`)
	s.Equal("foobar", db.Schema())
}

func TestDBSuite(t *testing.T) {
	RunTestSuite(t, new(DBSuite))
}
