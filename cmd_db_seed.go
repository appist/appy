package appy

import "os"

func newDbSeedCommand(config *Config, dbManager *DbManager, logger *Logger) *Cmd {
	var target string

	cmd := &Cmd{
		Use:   "db:seed",
		Short: "Seed the database(default: all, use --database to specify just 1) for the current environment",
		Run: func(cmd *Cmd, args []string) {
			if IsConfigErrored(config, logger) || IsDbManagerErrored(config, dbManager, logger) {
				os.Exit(-1)
			}

			if len(dbManager.dbs) < 1 {
				logger.Infof("No database is defined in pkg/config/.env.%s", config.AppyEnv)
				os.Exit(0)
			}

			runDbSeedAll(config, dbManager, logger, target)
		},
	}

	cmd.Flags().StringVar(&target, "database", "", "The target database to seed")
	return cmd
}

func runDbSeedAll(config *Config, dbManager *DbManager, logger *Logger, target string) {
	err := dbManager.ConnectAll(true)
	if err != nil {
		logger.Fatal(err)
	}
	defer dbManager.CloseAll()

	if target != "" {
		db := dbManager.Db(target)
		if db == nil {
			logger.Infof("No database called '%s' defined in pkg/config/.env.%s", target, config.AppyEnv)
			os.Exit(0)
		}

		if db.config.Replica {
			logger.Infof("Unable to run seeding on '%s' database that is a replica", target)
			os.Exit(0)
		}

		logger.Infof("Seeding '%s' database...", target)
		err := db.Seed()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Seeding '%s' database... DONE", target)

		return
	}

	for name, db := range dbManager.dbs {
		if db.config.Replica {
			continue
		}

		logger.Infof("Seeding '%s' database...", name)
		err := db.Seed()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Infof("Seeding '%s' database... DONE", name)
	}
}
