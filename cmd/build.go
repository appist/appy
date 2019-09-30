package cmd

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	ah "github.com/appist/appy/http"

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
			ssrPaths := []string{}
			for _, route := range s.GetAllRoutes() {
				if route.Method == "GET" {
					ssrPaths = append(ssrPaths, route.Path)
				}
			}

			logger.Info("Building the web app...")
			buildWebCmd := exec.Command("npm", "run", "build")
			buildWebCmd.Env = os.Environ()
			buildWebCmd.Env = append(buildWebCmd.Env, "APPY_SSR_PATHS="+strings.Join(ssrPaths, ","))
			buildWebCmd.Dir = "./web"
			buildWebCmd.Stdout = os.Stdout
			buildWebCmd.Stderr = os.Stderr
			if err := buildWebCmd.Run(); err != nil {
				logger.Fatal(err)
			}
			logger.Info("Building the web app... DONE")

			wd, err := os.Getwd()
			if err != nil {
				logger.Fatal(err)
			}

			binaryName := path.Base(wd)
			assetsPath := "assets"
			assetsPathForSSR := assetsPath + "/" + ah.SSRAssetsDir
			ssrSrcPath := "app"

			logger.Infof("Copying server-side assets from '%s' into '%s'...", ssrSrcPath+"/{locale,view}", assetsPathForSSR+"{locale,view}")
			err = copy.Copy(ssrSrcPath+"/view", assetsPathForSSR+"/view")
			if err != nil {
				logger.Fatal(err)
			}

			err = copy.Copy(ssrSrcPath+"/locale", assetsPathForSSR+"/locale")
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infof("Copying server-side assets from '%s' into '%s'... DONE", ssrSrcPath+"/{locale,view}", assetsPathForSSR+"{locale,view}")

			oldStdout := os.Stdout
			os.Stdout = nil

			logger.Infof("Compiling assets folder into main_assets.go...")
			err = vfsgen.Generate(http.Dir(assetsPath), vfsgen.Options{Filename: "main_assets.go", VariableName: "assets"})
			if err != nil {
				logger.Fatal(err)
			}
			logger.Infof("Compiling assets folder into main_assets.go... DONE")
			os.Stdout = oldStdout

			goPath, err := exec.LookPath("go")
			if err != nil {
				logger.Fatal(err)
			}

			logger.Info("Building the binary...")
			buildBinaryCmd := exec.Command(goPath, "build", "-a", "-tags", "netgo", "-ldflags", "-w -extldflags '-static' -X appist/appy/support.Build=release", "-o", binaryName, ".")
			buildBinaryCmd.Stderr = os.Stderr
			if err = buildBinaryCmd.Run(); err != nil {
				logger.Fatal(err)
			}
			logger.Info("Building the binary... DONE")

			_, err = exec.LookPath("upx")
			if err == nil {
				logger.Info("Compressing the binary with upx...")
				compressBinaryCmd := exec.Command("upx", binaryName)
				compressBinaryCmd.Stderr = os.Stderr
				if err = compressBinaryCmd.Run(); err != nil {
					logger.Fatal(err)
				}
				logger.Info("Compressing the binary with upx... DONE")
			}
		},
	}
}
