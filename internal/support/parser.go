package support

import (
	"reflect"
	"strings"

	"github.com/caarlos0/env"
)

// ParseEnv parses the environment variables into the config.
func ParseEnv(c interface{}) error {
	err := env.ParseWithFuncs(c, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(map[string]string{}): func(v string) (interface{}, error) {
			newMaps := map[string]string{}
			maps := strings.Split(v, ",")
			for _, m := range maps {
				splits := strings.Split(m, ":")
				if len(splits) != 2 {
					continue
				}

				newMaps[splits[0]] = splits[1]
			}

			return newMaps, nil
		},
		reflect.TypeOf([]byte{}): func(v string) (interface{}, error) {
			return []byte(v), nil
		},
		reflect.TypeOf([][]byte{}): func(v string) (interface{}, error) {
			newBytes := [][]byte{}
			bytes := strings.Split(v, ",")
			for _, b := range bytes {
				newBytes = append(newBytes, []byte(b))
			}

			return newBytes, nil
		},
	})

	if err != nil {
		return err
	}

	return nil
}
