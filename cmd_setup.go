package appy

import (
	"net/http"
	"os"
	"time"
)

func newSetupCommand(config *Config, dbManager *DbManager, logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "setup",
		Short: "Run dc:up, db:create, db:schema:load and db:seed",
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(config, logger) || IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("up", assets)
			if err != nil {
				logger.Fatal(err)
			}

			time.Sleep(5 * time.Second)
			runDbCreateAll(config, dbManager, logger)
			runDbSchemaLoad(config, dbManager, logger)
			runDbSeedAll(config, dbManager, logger, "")
		},
	}

	return cmd
}
