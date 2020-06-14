package record

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/bxcodec/faker/v3"
)

type (
	modelAssocSuite struct {
		test.Suite
		buffer    *bytes.Buffer
		dbManager *Engine
		i18n      *support.I18n
		logger    *support.Logger
		writer    *bufio.Writer
	}
)

func (s *modelAssocSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
	asset := support.NewAsset(nil, "testdata")
	config := support.NewConfig(asset, s.logger)
	s.i18n = support.NewI18n(asset, config, s.logger)
}

func (s *modelAssocSuite) TearDownTest() {
	for _, database := range s.dbManager.Databases() {
		s.Nil(database.Close())
	}
}

func (s *modelAssocSuite) model(v interface{}, opts ...ModelOption) Modeler {
	return NewModel(s.dbManager, v, opts...)
}

func (s *modelAssocSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s?multiStatements=true&parseTime=true", database))
		os.Setenv("DB_URI_REPLICA_PRIMARY_REPLICA", "true")
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
			os.Unsetenv("DB_URI_REPLICA_PRIMARY_REPLICA")
		}()

		query = `
CREATE TABLE IF NOT EXISTS authors (
	id INT AUTO_INCREMENT,
	name VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS books (
	id INT AUTO_INCREMENT,
	author_id INT,
	name VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	case "postgres":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
		os.Setenv("DB_URI_PRIMARY_REPLICA", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15433/%s?sslmode=disable&connect_timeout=5", database))
		os.Setenv("DB_URI_REPLICA_PRIMARY_REPLICA", "true")
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
			os.Unsetenv("DB_URI_REPLICA_PRIMARY_REPLICA")
		}()

		query = `
CREATE TABLE IF NOT EXISTS authors (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS books (
	id SERIAL PRIMARY KEY,
	author_id INT,
	name VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	updated_at TIMESTAMP
);
`
	}

	s.dbManager = NewEngine(s.logger, s.i18n)
	db := s.dbManager.DB("primary")

	err := db.DropDB(database)
	s.Nil(err)

	err = db.CreateDB(database)
	s.Nil(err)

	// Ensure master replication is completed.
	for _, database := range s.dbManager.Databases() {
		for true {
			if err := database.Connect(); err != nil {
				continue
			}

			if err := database.Ping(); err == nil {
				time.Sleep(100 * time.Millisecond)
				break
			}
		}
	}

	_, err = db.Exec(query)
	s.Nil(err)
}

func (s *modelAssocSuite) TestBelongsTo() {
	type authorM struct {
		Model     `masters:"primary" tableName:"authors" autoIncrement:"id" faker:"-"`
		ID        int64 `faker:"-"`
		Name      string
		CreatedAt support.ZTime `db:"created_at" faker:"-"`
		UpdatedAt support.ZTime `db:"updated_at" faker:"-"`
	}

	type bookM struct {
		Model     `masters:"primary" tableName:"books" autoIncrement:"id" faker:"-"`
		ID        int64 `faker:"-"`
		Name      string
		Author    *authorM      `association:"belongsTo" faker:"-"`
		AuthorID  int64         `db:"author_id"`
		CreatedAt support.ZTime `db:"created_at" faker:"-"`
		UpdatedAt support.ZTime `db:"updated_at" faker:"-"`
	}

	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_belongs_to_with_"+adapter)

		book := bookM{}
		s.Nil(faker.FakeData(&book))

		count, errs := s.model(&book).Create().Exec()
		s.Equal(int64(0), count)
		s.EqualError(errs[0], "Author cannot be nil")

		book = bookM{
			Author: &authorM{
				Name: "foo",
			},
			Name: "golang tutorial",
		}
		count, errs = s.model(&book).Create().Exec()
		s.Equal(0, len(errs))
		s.Equal(int64(1), count)

		author := book.Author
		book = bookM{
			Author: author,
			Name:   "nodejs tutorial",
		}
		count, errs = s.model(&book).Create().Exec()
		s.Equal(0, len(errs))
		s.Equal(int64(1), count)

		books := []bookM{
			{
				Author: &authorM{
					Name: "bar",
				},
				Name: "python tutorial",
			},
			{
				Author: author,
				Name:   "ruby tutorial",
			},
		}
		count, errs = s.model(&books).Create().Exec()
		s.Equal(0, len(errs))
		s.Equal(int64(2), count)
	}
}

func TestModelAssocSuite(t *testing.T) {
	test.Run(t, new(modelAssocSuite))
}
