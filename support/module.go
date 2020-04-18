package support

import (
	"io/ioutil"
	"os"
	"strings"
)

// ModuleName parses go.mod and return the module name.
func ModuleName() string {
	modulePrefix := "module "
	wd, _ := os.Getwd()
	data, _ := ioutil.ReadFile(wd + "/go.mod")

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, modulePrefix) {
			module := strings.TrimPrefix(line, modulePrefix)
			return strings.Trim(module, "\n")
		}
	}

	return ""
}
