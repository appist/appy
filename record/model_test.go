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
		Modeler   `masters:"primary" replicas:"" tableName:"admins" primaryKeys:"id" faker:"-"`
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

	MasterOnlyUser struct {
		Modeler   `masters:"primary" replicas:"" tableName:"admins" primaryKeys:"id" faker:"-"`
		ID        int64      `db:"id" orm:"auto_increment:true" faker:"-"`
		Age       int64      `db:"-"`
		Email     string     `db:"email" faker:"email,unique"`
		Username  string     `db:"username" faker:"username,unique"`
		CreatedAt *time.Time `db:"created_at" faker:"-"`
		DeletedAt *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt *time.Time `db:"updated_at" faker:"-"`
	}

	ReplicaOnlyUser struct {
		Modeler   `masters:"" replicas:"primaryReplica" tableName:"admins" primaryKeys:"id" faker:"-"`
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
		os.Setenv("DB_URI_REPLICA_PRIMARY_REPLICA", "true")
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
			os.Unsetenv("DB_URI_REPLICA_PRIMARY_REPLICA")
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
		os.Setenv("DB_URI_REPLICA_PRIMARY_REPLICA", "true")
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
			os.Unsetenv("DB_URI_REPLICA_PRIMARY_REPLICA")
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

func (s *modelSuite) TestAll() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_all_with_"+adapter)

		users := []User{}
		for i := 0; i < 10; i++ {
			u := User{}
			s.Nil(faker.FakeData(&u))
			users = append(users, u)
		}
		count, err := s.model(&users).Create().Exec(nil, false)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		users = []User{}
		count, err = s.model(&users).All().Exec(nil, false)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		// Wait for replication.
		time.Sleep(500 * time.Millisecond)
		users = []User{}
		count, err = s.model(&users).All().Exec(nil, true)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		for idx, u := range users {
			s.Equal(int64(idx+1), u.ID)
		}
	}
}

func (s *modelSuite) TestCount() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_count_with_"+adapter)

		users := []User{}
		for i := 0; i < 10; i++ {
			u := User{}
			s.Nil(faker.FakeData(&u))
			users = append(users, u)
		}

		count, err := s.model(&users).Create().Exec(nil, false)
		s.Equal(int64(10), count)
		s.Nil(err)

		for idx, u := range users {
			s.Equal(int64(idx+1), u.ID)
		}

		var user User
		count, err = s.model(&user).Count().Exec(nil, false)
		s.Equal(int64(10), count)
		s.Nil(err)

		count, err = s.model(&user).Select("DISTINCT concat(email, username)").Count().Exec(nil, false)
		s.Equal(int64(10), count)
		s.Nil(err)
	}
}

func (s *modelSuite) TestCreate() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_create_with_"+adapter)

		var user User
		s.Nil(faker.FakeData(&user))

		count, err := s.model(&user).Create().Exec(nil, false)
		s.Equal(int64(1), count)
		s.Equal(int64(1), user.ID)
		s.Nil(err)

		users := []User{}
		for i := 0; i < 10; i++ {
			u := User{}
			s.Nil(faker.FakeData(&u))
			users = append(users, u)
		}

		count, err = s.model(&users).Create().Exec(nil, false)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		for idx, u := range users {
			s.Equal(int64(idx+2), u.ID)
		}
	}
}

func (s *modelSuite) TestCustomTableName() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_custom_table_name_with_"+adapter)

		users := []AdminUser{}
		for i := 0; i < 10; i++ {
			u := AdminUser{}
			s.Nil(faker.FakeData(&u))
			users = append(users, u)
		}
		count, err := s.model(&users).Create().Exec(nil, false)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		users = []AdminUser{}
		count, err = s.model(&users).All().Exec(nil, false)
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		for idx, u := range users {
			s.Equal(int64(idx+1), u.ID)
		}
	}
}

func (s *modelSuite) TestEmptyQueryBuilder() {
	var user User
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Exec(nil, false)
	s.Equal(int64(0), count)
	s.Error(ErrModelEmptyQueryBuilder, err)
}

func (s *modelSuite) TestIgnoreTag() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_ignore_tag_"+adapter)

		var user User
		s.Nil(faker.FakeData(&user))
		s.NotContains(":age", s.model(&user).Create().SQL())
	}
}

func (s *modelSuite) TestMissingMasterDB() {
	var user ReplicaOnlyUser
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Create().Exec(nil, false)
	s.Equal(int64(0), count)
	s.Error(ErrModelMissingMasterDB, err)
}

func (s *modelSuite) TestMissingReplicaDB() {
	var user MasterOnlyUser
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Create().Exec(nil, true)
	s.Equal(int64(0), count)
	s.Error(ErrModelMissingReplicaDB, err)
}

func TestModelSuite(t *testing.T) {
	test.Run(t, new(modelSuite))
}
