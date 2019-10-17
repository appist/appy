package cmd

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/appist/appy/core"
	"github.com/otiai10/copy"
	"github.com/shurcooL/vfsgen"
)

// NewBuildCommand compiles the static assets into go files and build the release mode binary, only available for debug
// build.
func NewBuildCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "build",
		Short: "Compile the static assets into go files and build the release mode binary, only available for debug build.",
		Run: func(cmd *AppCmd, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				s.Logger.Fatal(err)
			}

			binaryName := path.Base(wd)
			assetsPath := "assets"
			assetsPathForSSR := assetsPath + "/" + s.SSRPaths["root"]
			os.RemoveAll(assetsPath)

			if _, err := os.Stat("./package.json"); !os.IsNotExist(err) {
				ssrPaths := []string{}
				for _, route := range s.Routes() {
					if route.Method == "GET" {
						ssrPaths = append(ssrPaths, route.Path)
					}
				}

				s.Logger.Info("Building the web app...")
				buildWebCmd := exec.Command("npm", "run", "build")
				buildWebCmd.Env = os.Environ()
				buildWebCmd.Env = append(buildWebCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
				buildWebCmd.Dir = wd
				buildWebCmd.Stdout = os.Stdout
				buildWebCmd.Stderr = os.Stderr
				if err := buildWebCmd.Run(); err != nil {
					s.Logger.Fatal(err)
				}
				s.Logger.Info("Building the web app... DONE")
			}

			s.Logger.Info("Copying server-side assets...")
			err = copy.Copy(s.SSRPaths["view"], assetsPathForSSR+"/"+s.SSRPaths["view"])
			if err != nil {
				s.Logger.Fatal(err)
			}

			err = copy.Copy(s.SSRPaths["locale"], assetsPathForSSR+"/"+s.SSRPaths["locale"])
			if err != nil {
				s.Logger.Fatal(err)
			}

			err = copy.Copy(s.SSRPaths["config"], assetsPathForSSR+"/"+s.SSRPaths["config"])
			if err != nil {
				s.Logger.Fatal(err)
			}

			keyFiles, _ := filepath.Glob(assetsPathForSSR + "/" + s.SSRPaths["config"] + "/*.key")
			for _, keyFile := range keyFiles {
				os.Remove(keyFile)
			}

			s.Logger.Info("Copying server-side assets... DONE")

			oldStdout := os.Stdout
			os.Stdout = nil

			generateMainAssets()
			s.Logger.Infof("Compiling assets folder into '%s'...", mainAssets)
			err = vfsgen.Generate(http.Dir(assetsPath), vfsgen.Options{Filename: mainAssets, VariableName: "assets"})
			if err != nil {
				s.Logger.Fatal(err)
			}
			s.Logger.Infof("Compiling assets folder into '%s'... DONE", mainAssets)
			os.Stdout = oldStdout

			// Add GraphQL/GRPC generator step
			goPath, err := exec.LookPath("go")
			if err != nil {
				s.Logger.Fatal(err)
			}

			s.Logger.Info("Building the binary...")
			buildBinaryCmd := exec.Command(goPath, "build", "-a", "-tags", "netgo", "-ldflags", "-w -extldflags '-static' -X github.com/appist/appy/core.Build=release", "-o", binaryName, ".")
			buildBinaryCmd.Stderr = os.Stderr
			if err = buildBinaryCmd.Run(); err != nil {
				s.Logger.Fatal(err)
			}
			s.Logger.Info("Building the binary... DONE")

			_, err = exec.LookPath("upx")
			if err == nil {
				s.Logger.Info("Compressing the binary with upx...")
				compressBinaryCmd := exec.Command("upx", binaryName)
				compressBinaryCmd.Stderr = os.Stderr
				if err = compressBinaryCmd.Run(); err != nil {
					s.Logger.Fatal(err)
				}
				s.Logger.Info("Compressing the binary with upx... DONE")
			}
		},
	}
}
