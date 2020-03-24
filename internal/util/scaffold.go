package util

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Scaffold generates a new project using the template.
func Scaffold(tplPath, name, description string) {
	err := filepath.Walk(tplPath,
		func(src string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			dest := strings.ReplaceAll(src, tplPath+"/", "")
			if info.IsDir() {
				err := os.MkdirAll(dest, 0777)
				if err != nil {
					return err
				}

				return nil
			}

			buf, err := ioutil.ReadFile(src)
			if err != nil {
				return err
			}

			file, err := os.Create(dest)
			if err != nil {
				return err
			}

			tpl, err := template.New("scaffold").Parse(string(buf))
			if err != nil {
				return err
			}

			err = tpl.Execute(file, map[string]string{
				"assetWelcomeCSS":         "{{assetPath(`styles/welcome.css`)}}",
				"blockHead":               "{{block head()}}",
				"blockBody":               "{{block body()}}",
				"blockEnd":                "{{end}}",
				"extendApplicationLayout": "{{extends \"../layouts/application.html\"}}",
				"projectName":             name,
				"projectDesc":             description,
				"translateWelcome":        "{{t(\"welcome\", `{\"Name\": \"John Doe\", \"Title\": \"` + t(\"title\") + `\"}`)}}",
				"yieldHead":               "{{yield head()}}",
				"yieldBody":               "{{yield body()}}",
			})

			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
