package util

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
					log.Fatal(err)
				}
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
