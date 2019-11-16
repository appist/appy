package cmd

import (
	"net/http"
	"os"
	"time"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewSetupCommand run dc:up, db:create, db:schema:load and db:seed.
func NewSetupCommand(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "setup",
		Short: "Run dc:up, db:create, db:schema:load and db:seed",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(config, logger) || appyorm.IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.Dbs()) < 1 {
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

func runDbSchemaLoad(config *appysupport.Config, dbManager *appyorm.DbManager, logger *appysupport.Logger) {
	logger.SetDbLogging(false)
	err := dbManager.ConnectAll(true)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbManager.CloseAll()

	for name, db := range dbManager.Dbs() {
		if db.Config().Replica {
			continue
		}

		logger.Infof("Loading schema for '%s' database...", name)
		_, err := db.Handle().Exec(db.Schema())
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Loading schema for '%s' database... DONE", name)
	}
}
