package appy

import (
	"net/http"
	"time"
)

func newSetupCommand(config *Config, dbManager *DbManager, logger *Logger, assets http.FileSystem) *Cmd {
	cmd := &Cmd{
		Use:   "setup",
		Short: "Run dc:up, db:create, db:schema:load and db:seed",
		Run: func(cmd *Cmd, args []string) {
			runDcUp(logger, assets)
			time.Sleep(5 * time.Second)
			runDbCreateAll(config, dbManager, logger)
			runDbSchemaLoad(config, dbManager, logger)
			runDbSeedAll(config, dbManager, logger, "")
		},
	}

	return cmd
}
