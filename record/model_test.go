package record

import (
	"bufio"
	"bytes"
	"context"
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

	User struct {
		Modeler    `masters:"primary" replicas:"primaryReplica" tableName:"" autoIncrement:"id" timezone:"local" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	AdminUser struct {
		Modeler    `masters:"primary" replicas:"" tableName:"admins" autoIncrement:"id" primaryKeys:"id,email" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	CallbackUser struct {
		Modeler    `masters:"primary" replicas:"primaryReplica" tableName:"" autoIncrement:"id" timezone:"local" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	DuplicateUser struct {
		Modeler    `masters:"primary" replicas:"primaryReplica" tableName:"" autoIncrement:"id" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	UserWithoutPK struct {
		Modeler    `masters:"primary" replicas:"primaryReplica" primaryKeys:"" tableName:"" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	MasterOnlyUser struct {
		Modeler    `masters:"primary" replicas:"" tableName:"admins" autoIncrement:"id" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}

	ReplicaOnlyUser struct {
		Modeler    `masters:"" replicas:"primaryReplica" tableName:"admins" faker:"-"`
		ID         int64      `db:"id" faker:"-"`
		Age        int64      `db:"-"`
		LoginCount *int64     `db:"login_count"`
		Email      string     `db:"email" faker:"email,unique"`
		Username   string     `db:"username" faker:"username,unique"`
		CreatedAt  *time.Time `db:"created_at" faker:"-"`
		DeletedAt  *time.Time `db:"deleted_at" faker:"-"`
		UpdatedAt  *time.Time `db:"updated_at" faker:"-"`
	}
)

func (u *CallbackUser) BeforeValidate() error {
	return nil
}

func (u *CallbackUser) AfterValidate() error {
	return nil
}

func (u *CallbackUser) BeforeCreate() error {
	return nil
}

func (u *CallbackUser) AfterCreate() error {
	return nil
}

func (u *CallbackUser) BeforeDelete() error {
	return nil
}

func (u *CallbackUser) AfterDelete() error {
	return nil
}

func (u *CallbackUser) BeforeFind() error {
	return nil
}

func (u *CallbackUser) AfterFind() error {
	return nil
}

func (u *CallbackUser) BeforeUpdate() error {
	return nil
}

func (u *CallbackUser) AfterUpdate() error {
	return nil
}

func (u *CallbackUser) AfterCommit() error {
	return nil
}

func (u *CallbackUser) AfterRollback() error {
	return nil
}

func (s *modelSuite) SetupTest() {
	s.logger, s.buffer, s.writer = support.NewTestLogger()
}

func (s *modelSuite) TearDownTest() {
	for _, database := range s.dbManager.Databases() {
		s.Nil(database.Close())
	}
}

func (s *modelSuite) model(v interface{}, opts ...ModelOption) Modeler {
	return NewModel(s.dbManager, v, opts...)
}

func (s *modelSuite) setupDB(adapter, database string) {
	var query string

	switch adapter {
	case "mysql":
		os.Setenv("DB_URI_PRIMARY", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13306/%s?multiStatements=true&parseTime=true", database))
		os.Setenv("DB_URI_PRIMARY_REPLICA", fmt.Sprintf("mysql://root:whatever@0.0.0.0:13307/%s?multiStatements=true&parseTime=true", database))
		os.Setenv("DB_URI_REPLICA_PRIMARY_REPLICA", "true")
		defer func() {
			os.Unsetenv("DB_URI_PRIMARY")
			os.Unsetenv("DB_URI_PRIMARY_REPLICA")
			os.Unsetenv("DB_URI_REPLICA_PRIMARY_REPLICA")
		}()

		query = `
CREATE TABLE IF NOT EXISTS admins (
	id INT AUTO_INCREMENT,
	login_count INT,
	email VARCHAR(64) UNIQUE NOT NULL,
	username VARCHAR(64) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP,
	PRIMARY KEY (id, email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	login_count INT,
	email VARCHAR(64) UNIQUE NOT NULL,
	username VARCHAR(64) UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS callback_users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	login_count INT,
	email VARCHAR(64) NOT NULL,
	username VARCHAR(64) NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS duplicate_users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	login_count INT,
	email VARCHAR(64) NOT NULL,
	username VARCHAR(64) NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS user_without_pks (
	id INT,
	login_count INT,
	email VARCHAR(64) NOT NULL,
	username VARCHAR(64) NOT NULL,
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
	login_count INT,
	email VARCHAR UNIQUE NOT NULL,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	login_count INT,
	email VARCHAR UNIQUE NOT NULL,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS callback_users (
	id SERIAL PRIMARY KEY,
	login_count INT,
	email VARCHAR NOT NULL,
	username VARCHAR NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS duplicate_users (
	id SERIAL PRIMARY KEY,
	login_count INT,
	email VARCHAR NOT NULL,
	username VARCHAR NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_without_pks (
	id SERIAL,
	login_count INT,
	email VARCHAR NOT NULL,
	username VARCHAR NOT NULL,
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

func (s *modelSuite) insertUsers() {
	users := []User{}
	for i := 0; i < 10; i++ {
		user := User{}
		s.Nil(faker.FakeData(&user))
		users = append(users, user)
	}

	count, err := s.model(&users).Create().Exec()
	s.Equal(10, len(users))
	s.Equal(int64(10), count)
	s.Nil(err)
}

func (s *modelSuite) TestAll() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_all_with_"+adapter)
		s.insertUsers()

		{
			var user User
			count, err := s.model(&user).All().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).All().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(1), users[0].ID)
			s.Nil(err)
		}

		{
			// Wait for replication.
			time.Sleep(500 * time.Millisecond)

			var users []User
			count, err := s.model(&users).All().Exec(ExecOption{UseReplica: true})
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			var user User
			count, err := s.model(&user).All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Equal(int64(1), users[0].ID)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			count, err := s.model(&user).All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}

		{
			var users []User
			count, err := s.model(&users).All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}
	}
}

func (s *modelSuite) TestAllTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_all_tx_with_"+adapter)
		s.insertUsers()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(1), users[0].ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			// Wait for replication.
			time.Sleep(500 * time.Millisecond)

			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec(ExecOption{UseReplica: true})
			s.Equal(int64(10), count)
			s.Equal(int64(1), users[0].ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Equal(int64(1), users[0].ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.All().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestCount() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_count_with_"+adapter)
		s.insertUsers()

		{
			var user User
			count, err := s.model(&user).Count().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Count().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		{
			var user User
			count, err := s.model(&user).Select("DISTINCT concat(email, username)").Count().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		{
			var user User
			count, err := s.model(&user).Where("$?").Count().Exec()
			s.Equal(int64(0), count)
			s.NotNil(err)
		}

		{
			user := User{}
			count, err := s.model(&user).Where("id > ?", 5).Count().Exec()
			s.Equal(int64(5), count)
			s.Nil(err)
		}

		{
			// Wait for replication.
			time.Sleep(500 * time.Millisecond)

			var user User
			count, err := s.model(&user).Count().Exec(ExecOption{UseReplica: true})
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			var user User
			count, err := s.model(&user).Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			count, err := s.model(&user).Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}

		{
			var users []User
			count, err := s.model(&users).Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}
	}
}

func (s *modelSuite) TestCountTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_count_tx_with_"+adapter)
		s.insertUsers()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := s.model(&user).Where("$?").Count().Exec()
			s.Equal(int64(0), count)
			s.NotNil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)
		}

		{
			var users []User
			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Count().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestCreate() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_create_with_"+adapter)

		{
			var user User
			s.Nil(faker.FakeData(&user))

			count, err := s.model(&user).Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			count, err := s.model(&users).Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			for idx, user := range users {
				s.Equal(int64(idx+2), user.ID)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			user := User{}
			s.Nil(faker.FakeData(&user))

			count, err := s.model(&user).Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Equal(int64(12), user.ID)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			count, err := s.model(&users).Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(10), count)
			s.Nil(err)

			for idx, user := range users {
				s.Equal(int64(idx+13), user.ID)
			}
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			user := User{}
			s.Nil(faker.FakeData(&user))

			count, err := s.model(&user).Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			count, err := s.model(&users).Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}
	}
}

func (s *modelSuite) TestCreateTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_create_tx_with_"+adapter)

		{
			var user User
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), user.ID)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(3), users[0].ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&users).Count().Exec()
			s.Equal(int64(11), count)
			s.Equal(int64(3), users[0].ID)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(13), users[0].ID)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&users).Count().Exec()
			s.Equal(int64(11), count)
			s.Equal(int64(13), users[0].ID)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			var user User
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Equal(int64(23), user.ID)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(12), count)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(12), count)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&users).Count().Exec()
			s.Equal(int64(12), count)
			s.Nil(err)
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.BeginContext(ctx, nil)
			s.NotNil(userModel.Tx())
			s.EqualError(err, "context deadline exceeded")
		}
	}
}

func (s *modelSuite) TestCustomTableName() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_custom_table_name_with_"+adapter)

		adminUsers := []AdminUser{}
		for i := 0; i < 10; i++ {
			adminUser := AdminUser{}
			s.Nil(faker.FakeData(&adminUser))
			adminUsers = append(adminUsers, adminUser)
		}

		count, err := s.model(&adminUsers).Create().Exec()
		s.Equal(10, len(adminUsers))
		s.Equal(int64(10), count)
		s.Nil(err)

		adminUsers = []AdminUser{}
		count, err = s.model(&adminUsers).All().Exec()
		s.Equal(10, len(adminUsers))
		s.Equal(int64(10), count)
		s.Nil(err)

		for idx, adminUser := range adminUsers {
			s.Equal(int64(idx+1), adminUser.ID)
		}
	}
}

func (s *modelSuite) TestDelete() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_delete_"+adapter)
		s.insertUsers()

		var users []User
		count, err := s.model(&users).All().Exec()
		s.Equal(10, len(users))
		s.Equal(int64(10), count)
		s.Nil(err)

		count, err = s.model(&users[5]).Delete().Exec()
		s.Equal(int64(1), count)
		s.Nil(err)

		count, err = s.model(&users[5]).Find().Exec()
		s.Equal(int64(0), count)
		s.Nil(err)

		users = []User{users[0], users[1]}
		count, err = s.model(&users).Delete().Exec()
		s.Equal(int64(2), count)
		s.Nil(err)

		users = []User{}
		count, err = s.model(&users).All().Exec()
		s.Equal(int64(7), count)
		s.Nil(err)
	}
}

func (s *modelSuite) TestDeleteTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_delete_tx_"+adapter)
		s.insertUsers()

		{
			user := User{ID: 1}
			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Delete().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			count, err = s.model(&User{ID: 1}).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			user := User{ID: 2}
			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Delete().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			count, err = s.model(&User{ID: 2}).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			users := []User{{ID: 2}, {ID: 3}}
			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Delete().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			count, err = s.model(&[]User{{ID: 2}, {ID: 3}}).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			users := []User{{ID: 4}, {ID: 5}}
			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Delete().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			count, err = s.model(&[]User{{ID: 4}, {ID: 5}}).Find().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestDeleteAll() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_delete_all_"+adapter)
		s.insertUsers()

		usersWithoutPK := []UserWithoutPK{}
		for i := 0; i < 10; i++ {
			u := UserWithoutPK{}
			s.Nil(faker.FakeData(&u))
			usersWithoutPK = append(usersWithoutPK, u)
		}
		count, err := s.model(&usersWithoutPK).Create().Exec()
		s.Equal(10, len(usersWithoutPK))
		s.Equal(int64(10), count)
		s.Nil(err)

		admins := []AdminUser{}
		for i := 0; i < 10; i++ {
			u := AdminUser{}
			s.Nil(faker.FakeData(&u))
			admins = append(admins, u)
		}
		count, err = s.model(&admins).Create().Exec()
		s.Equal(10, len(admins))
		s.Equal(int64(10), count)
		s.Nil(err)

		{
			user := User{}
			count, err = s.model(&user).Where("id ?").DeleteAll().Exec()
			s.Equal(int64(0), count)
			s.NotNil(err)
		}

		{
			user := User{}
			count, err = s.model(&user).Where("id = ?", 0).DeleteAll().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			user := User{ID: 9}
			count, err = s.model(&user).DeleteAll().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			count, err = s.model(&User{}).Count().Exec()
			s.Equal(int64(9), count)
			s.Nil(err)
		}

		{
			admin := AdminUser{ID: 9}
			count, err = s.model(&admin).DeleteAll().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			count, err = s.model(&AdminUser{}).Count().Exec()
			s.Equal(int64(9), count)
			s.Nil(err)
		}

		{
			admin := AdminUser{ID: 8, Email: "foo", Username: "bar"}
			count, err = s.model(&admin).DeleteAll().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)

			count, err = s.model(&AdminUser{}).Count().Exec()
			s.Equal(int64(9), count)
			s.Nil(err)

			admin = AdminUser{ID: 8, Email: admins[7].Email, Username: "bar"}
			count, err = s.model(&admin).DeleteAll().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			count, err = s.model(&AdminUser{}).Count().Exec()
			s.Equal(int64(8), count)
			s.Nil(err)
		}

		{
			admins := []AdminUser{
				{ID: 7, Email: admins[6].Email, Username: "bar"},
				{ID: 6, Email: admins[5].Email, Username: "bar"},
			}

			count, err = s.model(&admins).DeleteAll().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			count, err = s.model(&AdminUser{}).Count().Exec()
			s.Equal(int64(6), count)
			s.Nil(err)
		}

		{
			usersWithoutPK = []UserWithoutPK{}
			count, err = s.model(&usersWithoutPK).DeleteAll().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			count, err = s.model(&UserWithoutPK{}).Count().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			count, err = s.model(&User{}).Where("id IN (?)", []int64{1, 2, 3}).DeleteAll().Exec(ExecOption{Context: ctx})
			s.Equal(int64(3), count)
			s.Nil(err)

			user := User{}
			count, err = s.model(&user).Where("id IN (?)", []int64{1, 2, 3}).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)

			users := []User{}
			count, err = s.model(&users).Where("id IN (?)", []int64{1, 2, 3}).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)

			user = User{}
			count, err = s.model(&user).Where("id = ?", 5).Find().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(5), user.ID)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			user := User{}
			count, err = s.model(&user).Where("id ?").DeleteAll().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.NotNil(err)
		}
	}
}

func (s *modelSuite) TestDeleteAllTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_delete_all_tx_"+adapter)
		s.insertUsers()

		{
			user := User{ID: 1}
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.DeleteAll().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(9), count)
			s.Nil(err)
		}

		{
			user := User{ID: 2}
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.DeleteAll().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(9), count)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			user := User{ID: 2}
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.DeleteAll().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(8), count)
			s.Nil(err)
		}

		{
			user := User{ID: 3}
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.DeleteAll().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(8), count)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			user := User{ID: 3}
			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.DeleteAll().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Count().Exec()
			s.Equal(int64(8), count)
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestEmptyQueryBuilder() {
	var user User
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Exec()
	s.Equal(int64(0), count)
	s.Error(ErrModelEmptyQueryBuilder, err)
}

func (s *modelSuite) TestFind() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_find_"+adapter)
		s.insertUsers()

		{
			count, err := s.model(&User{}).Where("id > ?", 50).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			user := User{ID: 1}
			count, err := s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user User
			count, err := s.model(&user).Where("id ?").Find().Exec()
			s.Equal(int64(0), count)
			s.NotNil(err)
		}

		{
			var user User
			count, err := s.model(&user).Where("id != ?", 0).Order("id ASC").Limit(1).Offset(5).Find().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(6), user.ID)
			s.Nil(err)
		}

		{
			var user User
			count, err := s.model(&user).Where("id = ?", 0).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("id ?").Find().Exec()
			s.Equal(int64(0), count)
			s.NotNil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("id > ?", 5).Find().Exec()
			s.Equal(5, len(users))
			s.Equal(int64(5), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Select("username").Where("id > ?", 5).Find().Exec()
			s.Equal(5, len(users))
			s.Equal(int64(5), count)
			s.Equal("", users[0].Email)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("email = ? AND id IN (?) AND username = ?", "barfoo", []int64{5, 6, 7}, "foobar").Order("id ASC").Find().Exec()
			s.Equal(0, len(users))
			s.Equal(int64(0), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("id != ?", 0).Order("id DESC").Limit(1).Find().Exec()
			s.Equal(1, len(users))
			s.Equal(int64(1), count)
			s.Equal(int64(10), users[0].ID)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("id != ?", 0).Order("id ASC").Limit(1).Offset(5).Find().Exec()
			s.Equal(1, len(users))
			s.Equal(int64(1), count)
			s.Equal(int64(6), users[0].ID)
			s.Nil(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		{
			count, err := s.model(&User{}).Where("id = ?", 5).Find().Exec(ExecOption{Context: ctx})
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var users []User
			count, err := s.model(&users).Where("id IN (?)", []int64{5, 6, 7}).Order("id ASC").Find().Exec(ExecOption{Context: ctx})
			s.Equal(3, len(users))
			s.Equal(int64(3), count)
			s.Equal(int64(5), users[0].ID)
			s.Equal(int64(6), users[1].ID)
			s.Equal(int64(7), users[2].ID)
			s.Nil(err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		{
			var user User
			count, err := s.model(&user).Where("id != ?", 0).Order("id ASC").Limit(1).Offset(5).Find().Exec(ExecOption{Context: ctx})
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}

		{
			var users []User
			count, err := s.model(&users).Where("id != ?", 0).Order("id ASC").Limit(1).Offset(5).Find().Exec(ExecOption{Context: ctx})
			s.Equal(0, len(users))
			s.Equal(int64(0), count)
			s.EqualError(err, "context deadline exceeded")
		}
	}
}

func (s *modelSuite) TestFindTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_find_tx_"+adapter)

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			count, err = model.Count().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			count, err = model.Count().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), user.ID)
			s.Nil(err)

			count, err = model.Count().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			count, err = model.Count().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
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

func (s *modelSuite) TestMissingMasterDB() {
	var user ReplicaOnlyUser
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Create().Exec()
	s.Equal(int64(0), count)
	s.Error(ErrModelMissingMasterDB, err)
}

func (s *modelSuite) TestMissingReplicaDB() {
	var user MasterOnlyUser
	s.Nil(faker.FakeData(&user))

	count, err := s.model(&user).Create().Exec(ExecOption{UseReplica: true})
	s.Equal(int64(0), count)
	s.Error(ErrModelMissingReplicaDB, err)
}

func (s *modelSuite) TestScan() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_scan_"+adapter)
		s.insertUsers()

		type customResult struct {
			ID    int64
			Total int64
		}

		{
			var (
				user   User
				result customResult
			)

			count, err := s.model(&user).Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 5).Order("id ASC").Limit(1).Offset(1).Scan(&result).Exec()
			s.Equal(int64(7), result.ID)
			s.Equal(int64(14), result.Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var (
				users   []User
				results []customResult
			)

			count, err := s.model(&users).Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 5).Order("id ASC").Limit(1).Offset(1).Scan(&results).Exec()
			s.Equal(1, len(results))
			s.Equal(int64(7), results[0].ID)
			s.Equal(int64(14), results[0].Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user, scanUser User
			count, err := s.model(&user).Where("id != ?", 0).Group("id").Having("id > ?", 5).Order("id ASC").Scan(&scanUser).Exec()
			s.Equal(int64(6), scanUser.ID)
			s.Equal(int64(1), count)
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestScanTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_scan_tx_"+adapter)

		type customResult struct {
			ID    int64
			Total int64
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			var result customResult
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id ASC").Limit(1).Offset(0).Scan(&result).Exec()
			s.Equal(int64(1), result.ID)
			s.Equal(int64(2), result.Total)
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			result = customResult{}
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id ASC").Limit(1).Offset(0).Scan(&result).Exec()
			s.Equal(int64(1), result.ID)
			s.Equal(int64(2), result.Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), user.ID)
			s.Nil(err)

			var result customResult
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&result).Exec()
			s.Equal(int64(2), result.ID)
			s.Equal(int64(4), result.Total)
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			result = customResult{}
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&result).Exec()
			s.Equal(int64(1), result.ID)
			s.Equal(int64(2), result.Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			var results []customResult
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&results).Exec()
			s.Equal(1, len(results))
			s.Equal(int64(12), results[0].ID)
			s.Equal(int64(24), results[0].Total)
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			results = []customResult{}
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&results).Exec()
			s.Equal(1, len(results))
			s.Equal(int64(12), results[0].ID)
			s.Equal(int64(24), results[0].Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			var results []customResult
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&results).Exec()
			s.Equal(1, len(results))
			s.Equal(int64(22), results[0].ID)
			s.Equal(int64(44), results[0].Total)
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			results = []customResult{}
			count, err = model.Select("id, SUM(id * 2) AS total").Where("id != ?", 0).Group("id").Having("id > ?", 0).Order("id DESC").Limit(1).Offset(0).Scan(&results).Exec()
			s.Equal(1, len(results))
			s.Equal(int64(12), results[0].ID)
			s.Equal(int64(24), results[0].Total)
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			var user, scanUser User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Where("id != ?", 0).Group("id").Having("id > ?", 5).Order("id DESC").Scan(&scanUser).Exec()
			s.Equal(int64(12), scanUser.ID)
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			scanUser = User{}
			count, err = model.Where("id != ?", 0).Group("id").Having("id > ?", 5).Order("id DESC").Scan(&scanUser).Exec()
			s.Equal(int64(12), scanUser.ID)
			s.Equal(int64(1), count)
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestShareTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_share_tx_"+adapter)

		{
			var du DuplicateUser
			s.Nil(faker.FakeData(&du))

			duModel := s.model(&du)
			err := duModel.Begin()
			s.Nil(err)

			count, err := duModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), du.ID)
			s.Nil(err)

			var u User
			s.Nil(faker.FakeData(&u))

			uModel := s.model(&u, ModelOption{duModel.Tx()})
			count, err = uModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), u.ID)
			s.Nil(err)

			err = duModel.Commit()
			s.Nil(err)

			count, err = s.model(&du).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal(int64(1), du.ID)

			count, err = s.model(&u).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal(int64(1), u.ID)
		}

		{
			var du DuplicateUser
			s.Nil(faker.FakeData(&du))

			duModel := s.model(&du)
			err := duModel.Begin()
			s.Nil(err)

			count, err := duModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), du.ID)
			s.Nil(err)

			var u User
			s.Nil(faker.FakeData(&u))

			uModel := s.model(&u, ModelOption{duModel.Tx()})
			count, err = uModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), u.ID)
			s.Nil(err)

			err = duModel.Rollback()
			s.Nil(err)

			du = DuplicateUser{}
			count, err = s.model(&du).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal(int64(1), du.ID)

			u = User{}
			count, err = s.model(&u).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal(int64(1), u.ID)
		}
	}
}

func (s *modelSuite) TestUpdate() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_update_"+adapter)

		{
			var user User
			s.Nil(faker.FakeData(&user))

			count, err := s.model(&user).Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			user.Email = "foo@gmail.com"
			user.Username = "foo"
			user.LoginCount = nil

			count, err = s.model(&user).Update().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			user = User{}
			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal("foo@gmail.com", user.Email)
			s.Equal("foo", user.Username)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			count, err := s.model(&users).Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			users[0].Email = "bar@gmail.com"
			users[0].Username = "bar"
			users[0].LoginCount = nil
			users[9].Email = "foobar@gmail.com"
			users[9].Username = "foobar"
			users[9].LoginCount = nil

			count, err = s.model(&users).Update().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			users = []User{}
			count, err = s.model(&users).All().Exec()
			s.Equal(int64(11), count)
			s.Nil(err)
			s.Equal("bar@gmail.com", users[1].Email)
			s.Equal("bar", users[1].Username)
			s.Equal("foobar@gmail.com", users[10].Email)
			s.Equal("foobar", users[10].Username)
		}
	}
}

func (s *modelSuite) TestUpdateTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_update_tx_"+adapter)

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			user.Email = "foo@gmail.com"
			user.Username = "foo"
			user.LoginCount = nil

			count, err = model.Update().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			user = User{}
			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal("foo@gmail.com", user.Email)
			s.Equal("foo", user.Username)
		}

		{
			var user User
			s.Nil(faker.FakeData(&user))

			model := s.model(&user)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), user.ID)
			s.Nil(err)

			user.Email = "foo1@gmail.com"
			user.Username = "foo1"
			user.LoginCount = nil

			count, err = model.Update().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			user = User{}
			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			users[0].Email = "bar@gmail.com"
			users[0].Username = "bar"
			users[0].LoginCount = nil
			users[9].Email = "foobar@gmail.com"
			users[9].Username = "foobar"
			users[9].LoginCount = nil

			count, err = model.Update().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = model.Commit()
			s.Nil(err)

			users = []User{}
			count, err = s.model(&users).All().Exec()
			s.Equal(int64(11), count)
			s.Nil(err)
			s.Equal("bar@gmail.com", users[1].Email)
			s.Equal("bar", users[1].Username)
			s.Equal("foobar@gmail.com", users[10].Email)
			s.Equal("foobar", users[10].Username)
		}

		{
			users := []User{}
			for i := 0; i < 10; i++ {
				user := User{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			model := s.model(&users)
			err := model.Begin()
			s.NotNil(model.Tx())
			s.Nil(err)

			count, err := model.Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			users[0].Email = "barfoo@gmail.com"
			users[0].Username = "barfoo"
			users[0].LoginCount = nil

			count, err = model.Update().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = model.Rollback()
			s.Nil(err)

			users = []User{}
			count, err = s.model(&users).All().Exec()
			s.Equal(int64(11), count)
			s.Nil(err)
		}
	}
}

func (s *modelSuite) TestUpdateAll() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_update_all_"+adapter)

		{
			var user DuplicateUser
			s.Nil(faker.FakeData(&user))

			count, err := s.model(&user).Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			count, err = s.model(&user).UpdateAll("email = ?, username = ?", "foo@gmail.com", "foo").Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal("foo@gmail.com", user.Email)
			s.Equal("foo", user.Username)
		}

		{
			users := []DuplicateUser{}
			for i := 0; i < 10; i++ {
				user := DuplicateUser{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			count, err := s.model(&users).Create().Exec()
			s.Equal(10, len(users))
			s.Equal(int64(10), count)
			s.Nil(err)

			users = []DuplicateUser{
				{ID: 1},
				{ID: 2},
			}
			count, err = s.model(&users).UpdateAll("email = ?, username = ?", "bar@gmail.com", "bar").Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			count, err = s.model(&users).Find().Exec()
			s.Equal(int64(2), count)
			s.Nil(err)

			s.Equal(int64(1), users[0].ID)
			s.Equal("bar@gmail.com", users[0].Email)
			s.Equal("bar", users[0].Username)
			s.Equal(int64(2), users[1].ID)
			s.Equal("bar@gmail.com", users[1].Email)
			s.Equal("bar", users[1].Username)
		}
	}
}

func (s *modelSuite) TestUpdateAllTx() {
	for _, adapter := range support.SupportedDBAdapters {
		s.setupDB(adapter, "test_model_update_all_tx_"+adapter)

		{
			var user DuplicateUser
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(1), user.ID)
			s.Nil(err)

			count, err = userModel.UpdateAll("email = ?, username = ?", "foo@gmail.com", "foo").Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(1), count)
			s.Nil(err)
			s.Equal("foo@gmail.com", user.Email)
			s.Equal("foo", user.Username)
		}

		{
			var user DuplicateUser
			s.Nil(faker.FakeData(&user))

			userModel := s.model(&user)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(1), count)
			s.Equal(int64(2), user.ID)
			s.Nil(err)

			count, err = userModel.UpdateAll("email = ?, username = ?", "foo@gmail.com", "foo").Exec()
			s.Equal(int64(1), count)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&user).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
			s.NotEqual("foo@gmail.com", user.Email)
			s.NotEqual("foo", user.Username)
		}

		{
			users := []DuplicateUser{}
			for i := 0; i < 10; i++ {
				user := DuplicateUser{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(3), users[0].ID)
			s.Nil(err)

			count, err = userModel.UpdateAll("email = ?, username = ?", "foo@gmail.com", "foo").Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Commit()
			s.Nil(err)

			count, err = s.model(&users).Find().Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			for i := 0; i < 10; i++ {
				s.Equal("foo@gmail.com", users[i].Email)
				s.Equal("foo", users[i].Username)
			}
		}

		{
			users := []DuplicateUser{}
			for i := 0; i < 10; i++ {
				user := DuplicateUser{}
				s.Nil(faker.FakeData(&user))
				users = append(users, user)
			}

			userModel := s.model(&users)
			err := userModel.Begin()
			s.NotNil(userModel.Tx())
			s.Nil(err)

			count, err := userModel.Create().Exec()
			s.Equal(int64(10), count)
			s.Equal(int64(13), users[0].ID)
			s.Nil(err)

			count, err = userModel.UpdateAll("email = ?, username = ?", "foo@gmail.com", "foo").Exec()
			s.Equal(int64(10), count)
			s.Nil(err)

			err = userModel.Rollback()
			s.Nil(err)

			count, err = s.model(&users).Find().Exec()
			s.Equal(int64(0), count)
			s.Nil(err)
		}
	}
}

func TestModelSuite(t *testing.T) {
	test.Run(t, new(modelSuite))
}
