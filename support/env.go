package support

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/caarlos0/env"
)

// ParseEnv parses the environment variables into config struct.
func ParseEnv(c interface{}) error {
	if err := env.ParseWithFuncs(c, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf([]byte{}):            parseByteArray,
		reflect.TypeOf([][]byte{}):          parseByte2DArray,
		reflect.TypeOf(map[string]int{}):    parseMapStrInt,
		reflect.TypeOf(map[string]string{}): parseMapStrStr,
		reflect.TypeOf(http.SameSite(1)):    parseHTTPSameSite,
	}); err != nil {
		return err
	}

	return nil
}

func parseByteArray(v string) (interface{}, error) {
	return []byte(v), nil
}

func parseByte2DArray(v string) (interface{}, error) {
	newBytes := [][]byte{}
	bytes := strings.Split(v, ",")
	for _, b := range bytes {
		newBytes = append(newBytes, []byte(b))
	}

	return newBytes, nil
}

func parseHTTPSameSite(v string) (interface{}, error) {
	ss, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}

	return http.SameSite(ss), nil
}

func parseMapStrInt(v string) (interface{}, error) {
	newMaps := map[string]int{}
	maps := strings.Split(v, ",")
	for _, m := range maps {
		splits := strings.Split(m, ":")
		if len(splits) != 2 {
			continue
		}

		val, err := strconv.Atoi(splits[1])
		if err != nil {
			return nil, err
		}

		newMaps[splits[0]] = val
	}

	return newMaps, nil
}

func parseMapStrStr(v string) (interface{}, error) {
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
}
