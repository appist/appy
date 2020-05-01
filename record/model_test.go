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
	modelSuite struct {
		test.Suite
		buffer    *bytes.Buffer
		db        DBer
		dbManager *Engine
		logger    *support.Logger
		writer    *bufio.Writer
	}

	User struct {
		Modeler   `masters:"primary" replicas:"" tableName:"" primaryKeys:"id" faker:"-"`
		ID        int64      `db:"id" orm:"auto_increment:true" faker:"-"`
		Email     string     `db:"email" faker:"email,unique"`
		Username  string     `db:"username" faker:"username,unique"`
		CreatedAt *time.Time `db:"created_at" faker:"-"`
		DeletedAt *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt *time.Time `db:"updated_at" faker:"-"`
	}
)

func (s *modelSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *modelSuite) TearDownTest() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *modelSuite) model(v interface{}) Modeler {
	return NewModel(s.dbManager, v)
}

func (s *modelSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE IF NOT EXISTS users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	email VARCHAR(64) UNIQUE NOT NULL,
	username VARCHAR(64) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`
	case "postgres":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15432/%s?sslmode=disable&connect_timeout=5", database))
		defer os.Unsetenv("DB_URI_PRIMARY")

		query = `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email VARCHAR UNIQUE NOT NULL,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
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

func (s *modelSuite) TestAll() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_all_with_"+adapter)

		newUsers := []User{}
		for i := 0; i < 10; i++ {
			u := User{}
			s.Nil(faker.FakeData(&u))
			newUsers = append(newUsers, u)
		}
		s.Nil(s.model(&newUsers).Create().Exec(nil))

		var users []User
		s.Nil(s.model(&users).All().Exec(nil))

		for idx, u := range users {
			s.Equal(int64(idx+1), u.ID)
		}
	}
}

func (s *modelSuite) TestCreate() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_create_with_"+adapter)

		var user User
		s.Nil(faker.FakeData(&user))
		s.Nil(s.model(&user).Create().Exec(nil))
		s.Equal(int64(1), user.ID)

		users := []User{}
		for i := 0; i < 10; i++ {
			u := User{}
			s.Nil(faker.FakeData(&u))
			users = append(users, u)
		}
		s.Nil(s.model(&users).Create().Exec(nil))

		for idx, u := range users {
			s.Equal(int64(idx+2), u.ID)
		}
	}
}

func TestModelSuite(t *testing.T) {
	test.Run(t, new(modelSuite))
}
