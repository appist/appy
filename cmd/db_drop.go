package cmd

import (
	"os"

	"github.com/appist/appy/core"
)

// NewDbDropCommand drops all databases from app/config/.env.<APPY_ENV>.
func NewDbDropCommand(config core.AppConfig, dbMap map[string]*core.AppDb) *AppCmd {
	cmd := &AppCmd{
		Use:   "db:drop",
		Short: "Drops all databases from \"app/config/.env.<APPY_ENV>\".",
		Run: func(cmd *AppCmd, args []string) {
			checkProtectedEnvs(config)
			logger.Infof("Dropping databases from app/config/.env.%s...", config.AppyEnv)

			err := core.ConnectDb(dbMap, logger)
			if err != nil {
				logger.Fatal(err)
			}

			if len(dbMap) < 1 {
				logger.Infof("No database is defined in app/config/.env.%s.", config.AppyEnv)
				os.Exit(-1)
			}

			var msgs, errs []string
			for _, db := range dbMap {
				if db.Config.Replica {
					continue
				}

				tmpMsgs, tmpErrs := dbDrop(db)
				msgs = append(msgs, tmpMsgs...)
				errs = append(errs, tmpErrs...)
			}

			if len(errs) > 0 {
				for _, err := range errs {
					logger.Infof(err)
				}

				os.Exit(-1)
			}

			for _, msg := range msgs {
				logger.Info(msg)
			}

			core.CloseDb(dbMap)
		},
	}

	return cmd
}
