package appy

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/otiai10/copy"
	"github.com/shurcooL/vfsgen"
)

var (
	mainAssets = "pkg/app/assets.go"
)

func newBuildCommand(s *Server) *Cmd {
	return &Cmd{
		Use:   "build",
		Short: "Compile the static assets into go files and build the release mode binary (debug build only)",
		Run: func(cmd *Cmd, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				s.logger.Fatal(err)
			}

			binaryName := path.Base(wd)
			assetsPath := "assets"
			assetsPathForSSR := assetsPath + "/" + s.ssrPaths["root"]
			os.RemoveAll(assetsPath)

			if _, err := os.Stat("./package.json"); !os.IsNotExist(err) {
				ssrPaths := []string{}
				for _, route := range s.Routes() {
					if route.Method == "GET" {
						ssrPaths = append(ssrPaths, route.Path)
					}
				}

				s.logger.Info("Building the web app...")
				buildWebCmd := exec.Command("npm", "run", "build")
				buildWebCmd.Env = os.Environ()
				buildWebCmd.Env = append(buildWebCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
				buildWebCmd.Dir = wd
				buildWebCmd.Stdout = os.Stdout
				buildWebCmd.Stderr = os.Stderr
				if err := buildWebCmd.Run(); err != nil {
					s.logger.Fatal(err)
				}
				s.logger.Info("Building the web app... DONE")
			}

			s.logger.Info("Copying server-side assets...")
			err = copy.Copy(s.ssrPaths["docker"], assetsPathForSSR+"/"+s.ssrPaths["docker"])
			if err != nil {
				s.logger.Fatal(err)
			}

			err = copy.Copy(s.ssrPaths["view"], assetsPathForSSR+"/"+s.ssrPaths["view"])
			if err != nil {
				s.logger.Fatal(err)
			}

			err = copy.Copy(s.ssrPaths["locale"], assetsPathForSSR+"/"+s.ssrPaths["locale"])
			if err != nil {
				s.logger.Fatal(err)
			}

			err = copy.Copy(s.ssrPaths["config"], assetsPathForSSR+"/"+s.ssrPaths["config"])
			if err != nil {
				s.logger.Fatal(err)
			}

			keyFiles, _ := filepath.Glob(assetsPathForSSR + "/" + s.ssrPaths["config"] + "/*.key")
			for _, keyFile := range keyFiles {
				os.Remove(keyFile)
			}

			s.logger.Info("Copying server-side assets... DONE")
			oldStdout := os.Stdout
			os.Stdout = nil

			generateMainAssets(s.logger)
			s.logger.Infof("Compiling assets folder into '%s'...", mainAssets)
			err = vfsgen.Generate(http.Dir(assetsPath), vfsgen.Options{PackageName: "app", Filename: mainAssets, VariableName: "assets"})
			if err != nil {
				s.logger.Fatal(err)
			}
			s.logger.Infof("Compiling assets folder into '%s'... DONE", mainAssets)
			os.Stdout = oldStdout

			if _, err := os.Stat(wd + "pkg/graphql/schema.gql"); !os.IsNotExist(err) {
				s.logger.Info("Generating GraphQL boilerplate code...")
				err = generateGQL(s)
				if err != nil {
					s.logger.Fatal(err.Error())
				}

				s.logger.Info("Generating GraphQL boilerplate code... DONE")
			}

			s.logger.Info("Building the binary...")
			goPath, err := exec.LookPath("go")
			if err != nil {
				s.logger.Fatal(err)
			}

			buildBinaryCmd := exec.Command(goPath, "build", "-a", "-tags", "netgo", "-ldflags", "-w -extldflags '-static' -X github.com/appist/appy.Build=release", "-o", binaryName, ".")
			buildBinaryCmd.Stderr = os.Stderr
			if err = buildBinaryCmd.Run(); err != nil {
				s.logger.Fatal(err)
			}
			fi, _ := os.Stat(binaryName)
			s.logger.Infof("Building the binary... DONE (size: %s)", humanize.Bytes(uint64(fi.Size())))

			_, err = exec.LookPath("upx")
			if err == nil {
				s.logger.Info("Compressing the binary with upx...")
				compressBinaryCmd := exec.Command("upx", binaryName)
				compressBinaryCmd.Stderr = os.Stderr
				if err = compressBinaryCmd.Run(); err != nil {
					s.logger.Fatal(err)
				}
				fi, _ := os.Stat(binaryName)
				s.logger.Infof("Compressing the binary with upx... DONE (size: %s)", humanize.Bytes(uint64(fi.Size())))
			}
		},
	}
}

func generateMainAssets(logger *Logger) {
	os.Remove(mainAssets)

	template := []byte(`
// Generated by appy. DO NOT EDIT.
package main
import "net/http"
var assets http.FileSystem
`)
	err := ioutil.WriteFile(mainAssets, template, 0644)
	if err != nil {
		logger.Fatal(err)
	}
}
