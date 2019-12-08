package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type CloneSuite struct {
	test.Suite
}

func (s *CloneSuite) SetupTest() {
}

func (s *CloneSuite) TearDownTest() {
}

func (s *CloneSuite) TestDeepClone() {
	type User struct {
		Email string
		Name  string
	}

	type Employee struct {
		Email string
		Name  string
		Role  string
	}

	user := User{Email: "john_doe@gmail.com", Name: "John Doe"}
	employee := Employee{}
	err := DeepClone(&employee, &user)
	s.Nil(err)
	s.Equal("john_doe@gmail.com", employee.Email)
	s.Equal("John Doe", employee.Name)

	employees := []Employee{}
	err = DeepClone(&employees, &user)
	s.Nil(err)
	s.Equal(1, len(employees))
	s.Equal("john_doe@gmail.com", employees[0].Email)
	s.Equal("John Doe", employees[0].Name)

	users := []User{
		{Email: "john_doe1@gmail.com", Name: "John Doe 1"},
		{Email: "john_doe2@gmail.com", Name: "John Doe 2"},
	}
	employees = []Employee{}
	err = DeepClone(&employees, &users)
	s.Nil(err)
	s.Equal(2, len(employees))
	s.Equal("john_doe1@gmail.com", employees[0].Email)
	s.Equal("John Doe 1", employees[0].Name)
	s.Equal("john_doe2@gmail.com", employees[1].Email)
	s.Equal("John Doe 2", employees[1].Name)
}

func TestCloneSuite(t *testing.T) {
	test.RunSuite(t, new(CloneSuite))
}
