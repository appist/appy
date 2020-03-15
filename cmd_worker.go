//+build !test

package appy

func newWorkerCommand(config *Config, dbManager *DBManager, logger *Logger, worker *Worker) *Command {
	return &Command{
		Use:   "worker",
		Short: "Run the worker to process background jobs",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			worker.Run()
		},
	}
}
