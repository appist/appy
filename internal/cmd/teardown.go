package cmd

import (
	"net/http"

	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewTeardownCommand run dc:down to teardown the docker-compose cluster.
func NewTeardownCommand(logger *appysupport.Logger, assets http.FileSystem) *Command {
	cmd := &Command{
		Use:   "teardown",
		Short: "Run dc:down to teardown the docker-compose cluster",
		Run: func(cmd *Command, args []string) {
			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("down", assets)
			if err != nil {
				logger.Fatal(err)
			}
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
