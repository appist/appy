package cmd

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
	"github.com/otiai10/copy"
	"github.com/shurcooL/vfsgen"
	"github.com/spf13/cobra"
)

// NewBuildCommand compiles the static assets into go files and build the release mode binary, only available for debug
// build.
func NewBuildCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Compile the static assets into go files and build the release mode binary, only available for debug build.",
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				support.Logger.Fatal(err)
			}

			binaryName := path.Base(wd)
			assetsPath := "assets"
			assetsPathForSSR := assetsPath + "/" + ah.SSRRootRelease
			os.RemoveAll(assetsPath)

			if _, err := os.Stat("./package.json"); !os.IsNotExist(err) {
				ssrPaths := []string{}
				for _, route := range s.Routes() {
					if route.Method == "GET" {
						ssrPaths = append(ssrPaths, route.Path)
					}
				}

				support.Logger.Info("Building the web app...")
				buildWebCmd := exec.Command("npm", "run", "build")
				buildWebCmd.Env = os.Environ()
				buildWebCmd.Env = append(buildWebCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
				buildWebCmd.Dir = wd
				buildWebCmd.Stdout = os.Stdout
				buildWebCmd.Stderr = os.Stderr
				if err := buildWebCmd.Run(); err != nil {
					support.Logger.Fatal(err)
				}
				support.Logger.Info("Building the web app... DONE")
			}

			support.Logger.Infof("Copying server-side assets from '%s' into '%s'...", ah.SSRRootDebug, assetsPathForSSR)
			err = copy.Copy(ah.SSRRootDebug+"/"+ah.SSRView, assetsPathForSSR+"/"+ah.SSRView)
			if err != nil {
				support.Logger.Fatal(err)
			}

			err = copy.Copy(ah.SSRRootDebug+"/"+ah.SSRLocale, assetsPathForSSR+"/"+ah.SSRLocale)
			if err != nil {
				support.Logger.Fatal(err)
			}

			err = copy.Copy(support.SSRConfig, assetsPathForSSR+"/"+support.SSRConfig)
			if err != nil {
				support.Logger.Fatal(err)
			}
			support.Logger.Infof("Copying server-side assets from '%s' into '%s'... DONE", ah.SSRRootDebug, assetsPathForSSR)

			oldStdout := os.Stdout
			os.Stdout = nil

			generateMainAssets()
			support.Logger.Infof("Compiling assets folder into '%s'...", mainAssets)
			err = vfsgen.Generate(http.Dir(assetsPath), vfsgen.Options{Filename: mainAssets, VariableName: "assets"})
			if err != nil {
				support.Logger.Fatal(err)
			}
			support.Logger.Infof("Compiling assets folder into '%s'... DONE", mainAssets)
			os.Stdout = oldStdout

			// Add GraphQL/GRPC generator step
			goPath, err := exec.LookPath("go")
			if err != nil {
				support.Logger.Fatal(err)
			}

			support.Logger.Info("Building the binary...")
			buildBinaryCmd := exec.Command(goPath, "build", "-a", "-tags", "netgo", "-ldflags", "-w -extldflags '-static' -X github.com/appist/appy/support.Build=release", "-o", binaryName, ".")
			buildBinaryCmd.Stderr = os.Stderr
			if err = buildBinaryCmd.Run(); err != nil {
				support.Logger.Fatal(err)
			}
			support.Logger.Info("Building the binary... DONE")

			_, err = exec.LookPath("upx")
			if err == nil {
				support.Logger.Info("Compressing the binary with upx...")
				compressBinaryCmd := exec.Command("upx", binaryName)
				compressBinaryCmd.Stderr = os.Stderr
				if err = compressBinaryCmd.Run(); err != nil {
					support.Logger.Fatal(err)
				}
				support.Logger.Info("Compressing the binary with upx... DONE")
			}
		},
	}
}
