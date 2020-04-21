package support

import (
	"regexp"
	"strings"
)

type inflection struct {
	regexp  *regexp.Regexp
	replace string
}

type inflectionRegular struct {
	find    string
	replace string
}

type inflectionIrregular struct {
	singular string
	plural   string
}

var (
	singularInflections = []inflectionRegular{
		{"s$", ""},
		{"(ss)$", "${1}"},
		{"(n)ews$", "${1}ews"},
		{"([ti])a$", "${1}um"},
		{"((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)(sis|ses)$", "${1}sis"},
		{"(^analy)(sis|ses)$", "${1}sis"},
		{"([^f])ves$", "${1}fe"},
		{"(hive)s$", "${1}"},
		{"(tive)s$", "${1}"},
		{"([lr])ves$", "${1}f"},
		{"([^aeiouy]|qu)ies$", "${1}y"},
		{"(s)eries$", "${1}eries"},
		{"(m)ovies$", "${1}ovie"},
		{"(c)ookies$", "${1}ookie"},
		{"(x|ch|ss|sh)es$", "${1}"},
		{"^(m|l)ice$", "${1}ouse"},
		{"(bus)(es)?$", "${1}"},
		{"(o)es$", "${1}"},
		{"(shoe)s$", "${1}"},
		{"(cris|test)(is|es)$", "${1}is"},
		{"^(a)x[ie]s$", "${1}xis"},
		{"(octop|vir)(us|i)$", "${1}us"},
		{"(alias|status)(es)?$", "${1}"},
		{"^(ox)en", "${1}"},
		{"(vert|ind)ices$", "${1}ex"},
		{"(matr)ices$", "${1}ix"},
		{"(quiz)zes$", "${1}"},
		{"(database)s$", "${1}"},
	}

	pluralInflections = []inflectionRegular{
		{"([a-z])$", "${1}s"},
		{"s$", "s"},
		{"^(ax|test)is$", "${1}es"},
		{"(octop|vir)us$", "${1}i"},
		{"(octop|vir)i$", "${1}i"},
		{"(alias|status)$", "${1}es"},
		{"(bu)s$", "${1}ses"},
		{"(buffal|tomat)o$", "${1}oes"},
		{"([ti])um$", "${1}a"},
		{"([ti])a$", "${1}a"},
		{"sis$", "ses"},
		{"(?:([^f])fe|([lr])f)$", "${1}${2}ves"},
		{"(hive)$", "${1}s"},
		{"([^aeiouy]|qu)y$", "${1}ies"},
		{"(x|ch|ss|sh)$", "${1}es"},
		{"(matr|vert|ind)(?:ix|ex)$", "${1}ices"},
		{"^(m|l)ouse$", "${1}ice"},
		{"^(m|l)ice$", "${1}ice"},
		{"^(ox)$", "${1}en"},
		{"^(oxen)$", "${1}"},
		{"(quiz)$", "${1}zes"},
	}

	irregularInflections = []inflectionIrregular{
		{"alumnus", "alumni"},
		{"analysis", "analyses"},
		{"appendix", "appendices"},
		{"automaton", "automata"},
		{"cactus", "cacti"},
		{"cafe", "cafes"},
		{"child", "children"},
		{"criterion", "criteria"},
		{"curriculum", "curricula"},
		{"datum", "data"},
		{"elf", "elves"},
		{"embargo", "embargoes"},
		{"focus", "foci"},
		{"foe", "foes"},
		{"foot", "feet"},
		{"fungus", "fungi"},
		{"goose", "geese"},
		{"graffito", "graffiti"},
		{"hero", "heroes"},
		{"index", "indices"},
		{"man", "men"},
		{"matrix", "matrices"},
		{"mombie", "mombies"},
		{"mouse", "mice"},
		{"move", "moves"},
		{"person", "people"},
		{"phenomenon", "phenomena"},
		{"radix", "radices"},
		{"schema", "schemata"},
		{"sex", "sexes"},
		{"stadium", "stadia"},
		{"stimulus", "stimuli"},
		{"stratum", "strata"},
		{"syllabus", "syllabi"},
		{"tooth", "teeth"},
		{"vertex", "vertices"},
		{"woman", "women"},
	}

	uncountableInflections = []string{
		"cash",
		"equipment",
		"evidence",
		"fish",
		"help",
		"information",
		"jeans",
		"luck",
		"money",
		"police",
		"progress",
		"rain",
		"research",
		"rice",
		"series",
		"software",
		"sheep",
		"species",
		"time",
		"traffic",
	}

	compiledPluralInflections   []inflection
	compiledSingularInflections []inflection
)

func init() {
	compile()
}

func compile() {
	compiledPluralInflections = []inflection{}
	compiledSingularInflections = []inflection{}

	for _, uncountable := range uncountableInflections {
		inf := inflection{
			regexp:  regexp.MustCompile("^(?i)(" + uncountable + ")$"),
			replace: "${1}",
		}
		compiledPluralInflections = append(compiledPluralInflections, inf)
		compiledSingularInflections = append(compiledSingularInflections, inf)
	}

	for _, value := range irregularInflections {
		infs := []inflection{
			{regexp: regexp.MustCompile(strings.ToUpper(value.singular) + "$"), replace: strings.ToUpper(value.plural)},
			{regexp: regexp.MustCompile(strings.Title(value.singular) + "$"), replace: strings.Title(value.plural)},
			{regexp: regexp.MustCompile(value.singular + "$"), replace: value.plural},
		}
		compiledPluralInflections = append(compiledPluralInflections, infs...)
	}

	for _, value := range irregularInflections {
		infs := []inflection{
			{regexp: regexp.MustCompile(strings.ToUpper(value.plural) + "$"), replace: strings.ToUpper(value.singular)},
			{regexp: regexp.MustCompile(strings.Title(value.plural) + "$"), replace: strings.Title(value.singular)},
			{regexp: regexp.MustCompile(value.plural + "$"), replace: value.singular},
		}
		compiledSingularInflections = append(compiledSingularInflections, infs...)
	}

	for i := len(pluralInflections) - 1; i >= 0; i-- {
		value := pluralInflections[i]
		infs := []inflection{
			{regexp: regexp.MustCompile(strings.ToUpper(value.find)), replace: strings.ToUpper(value.replace)},
			{regexp: regexp.MustCompile(value.find), replace: value.replace},
			{regexp: regexp.MustCompile("(?i)" + value.find), replace: value.replace},
		}
		compiledPluralInflections = append(compiledPluralInflections, infs...)
	}

	for i := len(singularInflections) - 1; i >= 0; i-- {
		value := singularInflections[i]
		infs := []inflection{
			{regexp: regexp.MustCompile(strings.ToUpper(value.find)), replace: strings.ToUpper(value.replace)},
			{regexp: regexp.MustCompile(value.find), replace: value.replace},
			{regexp: regexp.MustCompile("(?i)" + value.find), replace: value.replace},
		}
		compiledSingularInflections = append(compiledSingularInflections, infs...)
	}
}

// Plural converts a word to its plural form.
func Plural(str string) string {
	for _, inflection := range compiledPluralInflections {
		if inflection.regexp.MatchString(str) {
			return inflection.regexp.ReplaceAllString(str, inflection.replace)
		}
	}

	return str
}

// Singular converts a word to its singular form.
func Singular(str string) string {
	for _, inflection := range compiledSingularInflections {
		if inflection.regexp.MatchString(str) {
			return inflection.regexp.ReplaceAllString(str, inflection.replace)
		}
	}

	return str
}
