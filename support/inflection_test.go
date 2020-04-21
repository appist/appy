package support

import (
	"strings"
	"testing"

	"github.com/appist/appy/test"
)

type inflectionSuite struct {
	test.Suite
	testCases map[string]string
}

func (s *inflectionSuite) SetupTest() {
	s.testCases = map[string]string{
		"alumnus":     "alumni",
		"star":        "stars",
		"STAR":        "STARS",
		"Star":        "Stars",
		"bus":         "buses",
		"fish":        "fish",
		"mouse":       "mice",
		"query":       "queries",
		"ability":     "abilities",
		"agency":      "agencies",
		"movie":       "movies",
		"archive":     "archives",
		"index":       "indices",
		"wife":        "wives",
		"safe":        "saves",
		"half":        "halves",
		"move":        "moves",
		"salesperson": "salespeople",
		"person":      "people",
		"spokesman":   "spokesmen",
		"man":         "men",
		"woman":       "women",
		"basis":       "bases",
		"diagnosis":   "diagnoses",
		"diagnosis_a": "diagnosis_as",
		"datum":       "data",
		"medium":      "media",
		"stadium":     "stadia",
		"analysis":    "analyses",
		"node_child":  "node_children",
		"child":       "children",
		"experience":  "experiences",
		"day":         "days",
		"comment":     "comments",
		"foobar":      "foobars",
		"newsletter":  "newsletters",
		"old_news":    "old_news",
		"news":        "news",
		"series":      "series",
		"species":     "species",
		"quiz":        "quizzes",
		"perspective": "perspectives",
		"ox":          "oxen",
		"photo":       "photos",
		"buffalo":     "buffaloes",
		"tomato":      "tomatoes",
		"dwarf":       "dwarves",
		"elf":         "elves",
		"information": "information",
		"equipment":   "equipment",
		"criterion":   "criteria",
	}
}

func (s *inflectionSuite) TearDownTest() {
}

func (s *inflectionSuite) TestPlural() {
	for key, value := range s.testCases {
		s.Equal(strings.ToUpper(value), Plural(strings.ToUpper(key)))
		s.Equal(strings.Title(value), Plural(strings.Title(key)))
		s.Equal(value, Plural(key))
	}

	s.Equal("", Plural(""))
}

func (s *inflectionSuite) TestSingular() {
	for key, value := range s.testCases {
		s.Equal(strings.ToUpper(key), Singular(strings.ToUpper(value)))
		s.Equal(strings.Title(key), Singular(strings.Title(value)))
		s.Equal(key, Singular(value))
	}

	s.Equal("", Singular(""))
}

func TestInflectionSuite(t *testing.T) {
	test.Run(t, new(inflectionSuite))
}
