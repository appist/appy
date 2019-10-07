package main

import (
	"log"
	"net/http"
	"path"
	"runtime"

	"github.com/shurcooL/vfsgen"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	currDir := path.Join(path.Dir(filename))
	err := vfsgen.Generate(http.Dir(currDir+"/../../tools/dist"), vfsgen.Options{
		Filename:     currDir + "/../../tools/assets.go",
		PackageName:  "tools",
		VariableName: "Assets",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
