package support

import (
	"net/http"
	"testing"

	"github.com/appist/appy/test"
)

type envSuite struct {
	test.Suite
}

func (s *envSuite) TestParseByteArray() {
	val, err := parseByteArray("foobar")

	s.Nil(err)
	s.Equal([]byte("foobar"), val)
}

func (s *envSuite) TestParseByte2DArray() {
	val, err := parseByte2DArray("a,b,c")

	s.Nil(err)
	s.Equal(
		[][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("c"),
		}, val)
}

func (s *envSuite) TestParseHTTPSameSite() {
	{
		val, err := parseHTTPSameSite("foobar")

		s.Equal("strconv.Atoi: parsing \"foobar\": invalid syntax", err.Error())
		s.Nil(val)
	}

	{
		val, err := parseHTTPSameSite("3")

		s.Nil(err)
		s.Equal(http.SameSiteStrictMode, val)
	}
}

func (s *envSuite) TestParseMapStrInt() {
	{
		val, err := parseMapStrInt("a:1,b:abc,c:3")

		s.Equal("strconv.Atoi: parsing \"abc\": invalid syntax", err.Error())
		s.Nil(val)
	}

	{
		val, err := parseMapStrInt("a:1,b,c:3")

		s.Nil(err)
		s.Equal(
			map[string]int{
				"a": 1,
				"c": 3,
			},
			val,
		)
	}

	{
		val, err := parseMapStrInt("a:1,b:2,c:3")

		s.Nil(err)
		s.Equal(
			map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			val,
		)
	}
}

func (s *envSuite) TestParseMapStrStr() {
	{
		val, err := parseMapStrStr("a:a,b:2,c:c")

		s.Nil(err)
		s.Equal(
			map[string]string{
				"a": "a",
				"b": "2",
				"c": "c",
			},
			val,
		)
	}

	{
		val, err := parseMapStrStr("a:a,b,c:c")

		s.Nil(err)
		s.Equal(
			map[string]string{
				"a": "a",
				"c": "c",
			},
			val,
		)
	}

	{
		val, err := parseMapStrStr("a:a,b:b,c:c")

		s.Nil(err)
		s.Equal(
			map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
			val,
		)
	}
}

func TestEnvSuite(t *testing.T) {
	test.Run(t, new(envSuite))
}
