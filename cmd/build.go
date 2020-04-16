package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/support"
	"github.com/dustin/go-humanize"
	"github.com/otiai10/copy"
	"github.com/shurcooL/vfsgen"
)

func newBuildCommand(asset *support.Asset, logger *support.Logger, server *pack.Server) *Command {
	var (
		platform string
		static   bool
	)

	goExe := "go"
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}

	cmd := &Command{
		Use:   "build",
		Short: "Compile the static assets into go files and build the release build binary (only available in debug build)",
		Run: func(cmd *Command, args []string) {
			releasePath := "dist"
			os.RemoveAll(releasePath)

			platforms, err := getPlatforms()
			if err != nil {
				logger.Fatal(err)
			}

			if platform != "" && !support.ArrayContains(platforms, platform) {
				logger.Fatalf("the '%s' platform isn't supported, refer to `%s tool dist list` for the supported value", platform, goExe)
			}

			wd, err := os.Getwd()
			if err != nil {
				logger.Fatal(err)
			}

			err = buildWebApp(logger, server, wd)
			if err != nil {
				logger.Fatal(err)
			}

			err = copyAssetsFolder(asset, logger, releasePath)
			if err != nil {
				logger.Fatal(err)
			}

			err = copyWebAppBuild(logger, releasePath)
			if err != nil {
				logger.Fatal(err)
			}

			err = buildGraphQLBoilerplate(logger, wd)
			if err != nil {
				logger.Fatal(err)
			}

			err = buildCompressedBinary(logger, platform, static, wd)
			if err != nil {
				logger.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", fmt.Sprintf("The platform for the binary to run on, see `%s tool dist list` for full list", goExe))
	cmd.Flags().BoolVar(&static, "static", false, "Specify if the binary should statically be built")
	return cmd
}

func buildCompressedBinary(logger *support.Logger, platform string, static bool, wd string) error {
	name := path.Base(wd)

	logger.Info("Building the binary...")

	goExe := "go"
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}

	goPath, err := exec.LookPath(goExe)
	if err != nil {
		return err
	}

	buildCmdArgs := []string{"build", "-a", "-tags", "netgo jsoniter", "-ldflags", "-X github.com/appist/appy/support.Build=release -s -w"}
	if static {
		buildCmdArgs[len(buildCmdArgs)-1] += " -extldflags '-static'"
	}

	buildCmdEnv := []string{}
	splits := strings.Split(platform, "/")
	if len(splits) == 2 {
		buildCmdEnv = []string{fmt.Sprintf("GOOS=%s", splits[0]), fmt.Sprintf("GOARCH=%s", splits[1])}

		if splits[0] == "windows" {
			name += ".exe"
		}
	}

	buildCmd := exec.Command(goPath, append(buildCmdArgs, []string{"-o", name, "."}...)...)
	buildCmd.Env = os.Environ()
	buildCmd.Env = append(buildCmd.Env, buildCmdEnv...)

	buildCmd.Stderr = os.Stderr
	if err = buildCmd.Run(); err != nil {
		return err
	}
	fi, _ := os.Stat(name)

	logger.Infof("Building the binary... DONE (size: %s)", humanize.Bytes(uint64(fi.Size())))

	upxExe := "upx"
	if runtime.GOOS == "windows" {
		upxExe += ".exe"
	}

	_, err = exec.LookPath(upxExe)
	if err == nil {
		logger.Info("Compressing the binary with upx...")

		compressCmd := exec.Command("upx", name)
		compressCmd.Stderr = os.Stderr
		if err = compressCmd.Run(); err != nil {
			return err
		}

		fi, _ := os.Stat(name)
		logger.Infof("Compressing the binary with upx... DONE (size: %s)", humanize.Bytes(uint64(fi.Size())))
	}

	return nil
}

func buildGraphQLBoilerplate(logger *support.Logger, wd string) error {
	if _, err := os.Stat(wd + "/pkg/graphql/config.yml"); err != nil {
		return nil
	}

	logger.Info("Generating GraphQL boilerplate code...")

	err := generateGQL()
	if err != nil {
		return err
	}

	logger.Info("Generating GraphQL boilerplate code... DONE")
	return nil
}

func buildWebApp(logger *support.Logger, server *pack.Server, wd string) error {
	if _, err := os.Stat(wd + "/package.json"); err != nil {
		return err
	}

	ssrPaths := []string{}
	for _, route := range server.Routes() {
		if route.Method == "GET" {
			ssrPaths = append(ssrPaths, route.Path)
		}
	}

	logger.Info("Building the web app...")

	cmd := exec.Command("npm", "run", "build")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "APPY_SSR_ROUTES="+strings.Join(ssrPaths, ","))
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	logger.Info("Building the web app... DONE")
	return nil
}

func copyAssetsFolder(asset *support.Asset, logger *support.Logger, releasePath string) error {
	logger.Infof("Copying server-side assets into '%s' folder...", releasePath)
	toCopies := []string{asset.Layout().Config(), asset.Layout().Docker(), asset.Layout().Locale(), asset.Layout().View()}

	for _, toCopy := range toCopies {
		err := copy.Copy(toCopy, releasePath+"/"+toCopy)
		if err != nil {
			return err
		}
	}

	keyFiles, _ := filepath.Glob(releasePath + "/" + asset.Layout().Config() + "/*.key")
	for _, keyFile := range keyFiles {
		os.Remove(keyFile)
	}

	gitIgnoreFiles, _ := filepath.Glob(releasePath + "/**/.gitkeep")
	for _, gitIgnoreFile := range gitIgnoreFiles {
		os.Remove(gitIgnoreFile)
	}

	logger.Infof("Copying server-side assets into '%s' folder... DONE", releasePath)
	return nil
}

func copyWebAppBuild(logger *support.Logger, releasePath string) error {
	oldStdout := os.Stdout
	os.Stdout = nil
	defer func() { os.Stdout = oldStdout }()

	assetPath := "pkg/app/asset.go"
	logger.Infof("Compiling '%s' folder into '%s'...", releasePath, assetPath)

	generateAssetTemplate(assetPath, logger)
	err := vfsgen.Generate(http.Dir(releasePath), vfsgen.Options{PackageName: "app", Filename: assetPath, VariableName: "asset"})
	if err != nil {
		return err
	}

	logger.Infof("Compiling '%s' folder into '%s'... DONE", releasePath, assetPath)
	return nil
}

func generateAssetTemplate(assetPath string, logger *support.Logger) error {
	os.Remove(assetPath)

	template := []byte(`
// Generated by appy. DO NOT EDIT.
package app
import "net/http"
var assets http.FileSystem
`)
	err := ioutil.WriteFile(assetPath, template, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getPlatforms() ([]string, error) {
	var data []byte

	goExe := "go"
	if runtime.GOOS == "windows" {
		goExe += ".exe"
	}

	goPath, err := exec.LookPath(goExe)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(goPath, "tool", "dist", "list")
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	stdout, _ := cmd.StdoutPipe()
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	data, err = ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}
