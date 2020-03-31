//+build !test

package appy

func newWorkCommand(config *Config, dbManager *DBManager, logger *Logger, worker *Worker) *Command {
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

			for _, db := range dbManager.databases {
				err := db.Connect()
				if err != nil {
					logger.Fatal(err)
				}
			}

			worker.Run()
		},
	}
}
