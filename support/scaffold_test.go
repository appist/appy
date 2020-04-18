package support

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type scaffoldSuite struct {
	test.Suite
}

func (s *scaffoldSuite) TestScaffold() {
	scaffoldPath := "./tmp/scaffold"

	err := os.RemoveAll(scaffoldPath)
	s.Nil(err)

	err = os.MkdirAll(scaffoldPath, 0777)
	s.Nil(err)

	err = os.Chdir(scaffoldPath)
	s.Nil(err)

	err = ioutil.WriteFile("go.mod", []byte("module appist"), 0777)
	s.Nil(err)

	opts := ScaffoldOptions{
		DBAdapter:   "foobar",
		Description: "scaffold test",
	}
	err = Scaffold(opts)
	s.EqualError(err, "DBAdapter 'foobar' is not supported, only '[mysql postgres]' are supported")

	opts = ScaffoldOptions{
		Description: "scaffold test",
	}
	err = Scaffold(opts)
	s.Nil(err)
}

func TestScaffoldSuite(t *testing.T) {
	test.Run(t, new(scaffoldSuite))
}
