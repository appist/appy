package cmd

import (
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
	"github.com/appist/appy/worker"
)

func newWorkCommand(config *support.Config, dbManager *record.Engine, logger *support.Logger, worker *worker.Engine) *Command {
	return &Command{
		Use:   "work",
		Short: "Run the worker to process background jobs",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			for _, db := range dbManager.Databases() {
				err := db.Connect()
				if err != nil {
					logger.Fatal(err)
				}
			}

			worker.Run()
		},
	}
}
