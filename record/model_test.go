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
		dbManager *Engine
		logger    *support.Logger
		writer    *bufio.Writer
	}

	AdminUser struct {
		Modeler   `masters:"primary" replicas:"primaryReplica" tableName:"admins" primaryKeys:"id" faker:"-"`
		ID        int64      `db:"id" orm:"auto_increment:true" faker:"-"`
		Age       int64      `db:"-"`
		Email     string     `db:"email" faker:"email,unique"`
		Username  string     `db:"username" faker:"username,unique"`
		CreatedAt *time.Time `db:"created_at" faker:"-"`
		DeletedAt *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt *time.Time `db:"updated_at" faker:"-"`
	}

	User struct {
		Modeler   `masters:"primary" replicas:"primaryReplica" tableName:"" primaryKeys:"id" faker:"-"`
		ID        int64      `db:"id" orm:"auto_increment:true" faker:"-"`
		Age       int64      `db:"-"`
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
	for _, database := range s.dbManager.Databases() {
		s.Nil(database.Close())
	}
}

func (s *modelSuite) model(v interface{}) Modeler {
	return NewModel(s.dbManager, v)
}

func (s *modelSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s?multiStatements=true", database))
		os.Setenv("DB_URI_PRIMARY_REPLICA", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13307/%s?multiStatements=true", database))
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
		}()

		query = `
CREATE TABLE IF NOT EXISTS admins (
	id INT PRIMARY KEY AUTO_INCREMENT,
	email VARCHAR(64) UNIQUE NOT NULL,
	username VARCHAR(64) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

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
		os.Setenv("DB_URI_PRIMARY_REPLICA", fmt.Sprintf("postgresql://postgres:whatever@0.0.0.0:15433/%s?sslmode=disable&connect_timeout=5", database))
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
		}()

		query = `
CREATE TABLE IF NOT EXISTS admins (
	id SERIAL PRIMARY KEY,
	email VARCHAR UNIQUE NOT NULL,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);

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
	db := s.dbManager.DB("primary")

	err := db.DropDB(database)
	s.Nil(err)

	err = db.CreateDB(database)
	s.Nil(err)

	// Wait for database replication to complete.
	time.Sleep(1 * time.Second)

	for _, database := range s.dbManager.Databases() {
		s.Nil(database.Connect())
	}

	_, err = db.Exec(query)
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

func (s *modelSuite) TestCustomTableName() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_custom_table_name_with_"+adapter)

		newAdminUsers := []AdminUser{}
		for i := 0; i < 10; i++ {
			au := AdminUser{}
			s.Nil(faker.FakeData(&au))
			newAdminUsers = append(newAdminUsers, au)
		}
		s.Nil(s.model(&newAdminUsers).Create().Exec(nil))

		var adminUsers []AdminUser
		s.Nil(s.model(&adminUsers).All().Exec(nil))

		for idx, au := range adminUsers {
			s.Equal(int64(idx+1), au.ID)
		}
	}
}

func (s *modelSuite) TestIgnoreTag() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_ignore_tag_"+adapter)

		var user User
		s.Nil(faker.FakeData(&user))
		s.NotContains(":age", s.model(&user).Create().SQL())
	}
}

func TestModelSuite(t *testing.T) {
	test.Run(t, new(modelSuite))
}
