package record

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

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

type dbSuite struct {
	test.Suite
	buffer    *bytes.Buffer
	db        DBer
	dbManager *Engine
	logger    *support.Logger
	writer    *bufio.Writer
}

func (s *dbSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *dbSuite) TearDownTest() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *dbSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	case "postgres":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE users (
	username varchar(32) DEFAULT NULL
);
`
	}

	s.dbManager = NewEngine(s.logger)
	s.db = s.dbManager.DB("primary")

	err := s.db.DropDB(database)
	s.Nil(err)

	err = s.db.CreateDB(database)
	s.Nil(err)

	err = s.db.Connect()
	s.Nil(err)

	_, err = s.db.Exec(query)
	s.Nil(err)
}

func (s *dbSuite) TestConnect() {
	{
		os.Setenv("DB_URI_PRIMARY", "mysql://root:@0.0.0.0:13306/test_connect")
		defer os.Unsetenv("DB_URI_PRIMARY")

		configs, errs := parseDBConfig()

		s.Equal(0, len(errs))
		s.Equal(1, len(configs))

		db := NewDB(configs["primary"], s.logger)
		err := db.Connect()

		s.Contains(err.(*mysql.MySQLError).Message, "Access denied for user 'root'@")
	}

	{
		os.Setenv("DB_URI_PRIMARY", "postgres://postgres:@0.0.0.0:15432/test_connect?sslmode=disable&connect_timeout=5")
		defer os.Unsetenv("DB_URI_PRIMARY")

		configs, errs := parseDBConfig()

		s.Equal(0, len(errs))
		s.Equal(1, len(configs))

		db := NewDB(configs["primary"], s.logger)
		err := db.Connect()

		s.Equal("password authentication failed for user \"postgres\"", err.(*pq.Error).Message)
	}
}

func (s *dbSuite) TestGenerateMigration() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	oldMigratePath := migratePath
	migratePath = "tmp/" + migratePath
	defer func() { migratePath = oldMigratePath }()

	dbConfigs, errs := parseDBConfig()
	s.Equal(0, len(errs))
	s.Equal(1, len(dbConfigs))

	db := NewDB(dbConfigs["primary"], s.logger)

	// Test gen:migration without Tx.
	err := db.GenerateMigration("CreateUsers", "primary", false)
	s.Nil(err)

	files := []string{}
	_ = filepath.Walk(migratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err := ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigration(")
	s.Contains(string(content), "// Up migration")
	s.Contains(string(content), "// Down migration")
	s.Contains(string(content), "func(db appy.DB) error {")

	err = os.RemoveAll(migratePath)
	s.Nil(err)

	// Test gen:migration with Tx.
	err = db.GenerateMigration("CreateUsers", "primary", true)
	s.Nil(err)

	files = []string{}
	_ = filepath.Walk(migratePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	content, err = ioutil.ReadFile(files[0])
	s.Nil(err)
	s.Contains(string(content), "db.RegisterMigrationTx(")
	s.Contains(string(content), "// Up migration")
	s.Contains(string(content), "// Down migration")
	s.Contains(string(content), "func(db appy.Tx) error {")

	err = os.RemoveAll(migratePath)
	s.Nil(err)
}

func (s *dbSuite) TestTransaction() {
	tt := map[string]string{
		"mysql":    "INSERT INTO users VALUES(?);",
		"postgres": "INSERT INTO users VALUES($1);",
	}

	for adapter, query := range tt {
		s.setupDB(adapter, "test_transaction")

		tx, err := s.db.Begin()
		s.Nil(err)

		_, err = tx.Exec(query, "foobar")
		s.Nil(err)
		tx.Rollback()

		var count int
		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")
		s.Nil(err)
		s.Equal(0, count)

		tx, err = s.db.Begin()
		s.Nil(err)

		_, err = tx.Exec(query, "foobar")
		s.Nil(err)
		tx.Commit()

		ctx, cancel := context.WithCancel(context.Background())
		tx, err = s.db.BeginContext(ctx, nil)

		s.Nil(err)

		_, err = tx.Exec(query, "barfoo")
		cancel()

		s.Nil(err)
		s.NotNil(tx.Commit())

		err = s.db.Get(&count, "SELECT COUNT(*) FROM users;")

		s.Nil(err)
		s.Equal(1, count)
	}
}

func (s *dbSuite) TestMigrateSeedRollback() {
	oldMigratePath := migratePath
	defer func() { migratePath = oldMigratePath }()

	database := "test_migrate_seed_rollback"
	for _, adapter := range supportedAdapters {
		migratePath = "tmp/" + adapter + "/"
		s.setupDB(adapter, database)

		err := s.db.RegisterMigrationTx(
			func(tx Txer) error {
				_, err := tx.Exec(``)

				return err
			},
			func(tx Txer) error {
				_, err := tx.Exec(``)

				return err
			},
			"create_orders",
		)
		s.Equal("invalid filename '\"create_orders\"', a valid example: 20060102150405_create_users.go", err.Error())

		err = s.db.RegisterMigrationTx(
			func(tx Txer) error {
				_, err := tx.Exec(`CREATE orders`)

				return err
			},
			func(tx Txer) error {
				_, err := tx.Exec(`DROP TABLE IF EXISTS orders;`)

				return err
			},
			"20200201165238_create_orders",
		)
		s.Nil(err)

		err = s.db.Migrate()
		switch adapter {
		case "mysql":
			s.Equal("Error 1064: You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'orders' at line 1", err.Error())
		case "postgres":
			s.Equal("pq: syntax error at or near \"orders\"", err.Error())
		}

		// Reset the migrations.
		s.db.(*DB).migrations = []*Migration{}

		err = s.db.RegisterMigrationTx(
			func(tx Txer) error {
				_, err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS orders (
						id int NOT NULL,
						username VARCHAR(255) NOT NULL
					);
				`)

				return err
			},
			func(tx Txer) error {
				_, err := tx.Exec(`DROP TABLE IF EXISTS orders;`)

				return err
			},
			"20200201165238_create_orders",
		)
		s.Nil(err)

		createIndexQuery := `CREATE INDEX orders_on_username ON orders(username);`
		dropIndexQuery := `ALTER TABLE orders DROP INDEX orders_on_username;`
		if adapter == "postgres" {
			createIndexQuery = `CREATE INDEX CONCURRENTLY orders_on_username ON orders(username);`
			dropIndexQuery = `DROP INDEX orders_on_username;`
		}

		err = s.db.RegisterMigration(
			func(db DBer) error {
				_, err := db.Exec(createIndexQuery)
				return err
			},
			func(db DBer) error {
				_, err := db.Exec(dropIndexQuery)
				return err
			},
			"20200202165238_add_orders_on_username_index",
		)
		s.Nil(err)

		err = s.db.Migrate()
		s.Nil(err)

		// Test db:migrate:status after db:migrate.
		migrations, err := s.db.MigrateStatus()
		s.Nil(err)
		s.Equal(2, len(migrations))
		s.Equal("up", migrations[0][0])
		s.Equal("up", migrations[1][0])

		// Test db:schema:dump after db:migrate.
		err = s.db.DumpSchema(database)
		s.Nil(err)

		schemaPath := migratePath + "/" + database + "/schema.go"
		_, err = os.Stat(schemaPath)
		s.Equal(false, os.IsNotExist(err))

		buf, err := ioutil.ReadFile(schemaPath)
		s.Nil(err)
		s.Contains(string(buf), fmt.Sprintf("package %s", database))
		s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))

		switch adapter {
		case "mysql":
			s.Contains(string(buf), "CREATE TABLE orders (")
			s.Contains(string(buf), "KEY orders_on_username (username)")
		case "postgres":
			s.Contains(string(buf), "CREATE TABLE IF NOT EXISTS public.orders (")
			s.Contains(string(buf), "CREATE INDEX orders_on_username ON public.orders USING btree (username)")
		}

		s.Contains(string(buf), "('20200201165238'),")
		s.Contains(string(buf), "('20200202165238');")

		// Test db:seed.
		insertQuery := `INSERT INTO orders VALUES(?, ?);`
		if adapter == "postgres" {
			insertQuery = `INSERT INTO orders VALUES($1, $2);`
		}

		s.db.RegisterSeedTx(
			func(tx Txer) error {
				usernames := []string{"john doe", "james bond"}

				for idx, username := range usernames {
					_, err := tx.Exec(insertQuery, idx, username)
					if err != nil {
						return err
					}
				}

				return nil
			},
		)

		err = s.db.Seed()
		s.Nil(err)

		var count int
		err = s.db.Get(&count, "SELECT COUNT(*) FROM orders")
		s.Nil(err)
		s.Equal(2, count)

		// Test db:rollback.
		err = s.db.Rollback()
		s.Nil(err)

		migrations, err = s.db.MigrateStatus()
		s.Nil(err)
		s.Equal(2, len(migrations))
		s.Equal("up", migrations[0][0])
		s.Equal("down", migrations[1][0])

		err = s.db.Rollback()
		s.Nil(err)

		migrations, err = s.db.MigrateStatus()
		s.Nil(err)
		s.Equal(2, len(migrations))
		s.Equal("down", migrations[0][0])
		s.Equal("down", migrations[1][0])

		// Test db:schema:dump after db:rollback.
		err = s.db.DumpSchema(database)
		s.Nil(err)

		buf, err = ioutil.ReadFile(schemaPath)
		s.Nil(err)
		s.Contains(string(buf), fmt.Sprintf("package %s", database))
		s.Contains(string(buf), fmt.Sprintf(`db := app.DB("%s")`, database))

		switch adapter {
		case "mysql":
			s.NotContains(string(buf), "CREATE TABLE orders (")
			s.NotContains(string(buf), "KEY orders_on_username (username)")
		case "postgres":
			s.NotContains(string(buf), "CREATE TABLE IF NOT EXISTS public.orders (")
			s.NotContains(string(buf), "CREATE INDEX orders_on_username ON public.orders USING btree (username)")
		}

		s.NotContains(string(buf), "('20200201165238'),")
		s.NotContains(string(buf), "('20200202165238');")

		err = os.RemoveAll(migratePath)
		s.Nil(err)
	}
}

func (s *dbSuite) TestSetSchema() {
	os.Setenv("DB_URI_PRIMARY", "postgres://postgres:whatever@0.0.0.0:15432/postgres?sslmode=disable&connect_timeout=5")
	defer os.Unsetenv("DB_URI_PRIMARY")

	configs, errs := parseDBConfig()
	s.Equal(0, len(errs))
	s.Equal(1, len(configs))

	db := NewDB(configs["primary"], s.logger)
	s.Equal("", db.Schema())

	db.SetSchema(`foobar`)
	s.Equal("foobar", db.Schema())
}

func TestDBSuite(t *testing.T) {
	test.Run(t, new(dbSuite))
}
