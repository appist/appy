package cmd

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newSetupCommand(asset *support.Asset, config *support.Config, dbManager *record.Engine, logger *support.Logger) *Command {
	return &Command{
		Use:   "setup",
		Short: "Run dc:up/db:create/db:schema:load/db:seed to setup the datastore with seed data",
		Run: func(cmd *Command, args []string) {
			if len(config.Errors()) > 0 {
				logger.Fatal(config.Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			err := checkDocker()
			if err != nil {
				logger.Fatal(err)
			}

			err = runDockerCompose("up", asset)
			if err != nil {
				logger.Fatal(err)
			}

			fmt.Println("Waiting for databases to be ready...")

			for _, db := range dbManager.Databases() {
				err := pingDB(db.Config().Host, db.Config().Port)
				for err != nil {
					time.Sleep(5 * time.Second)
					err = pingDB(db.Config().Host, db.Config().Port)
				}
			}

			fmt.Println("Waiting for databases to be ready... DONE")

			runDBCreateAll(config, dbManager, logger)
			runDBSchemaLoad(config, dbManager, logger)
			runDBSeedAll(config, dbManager, logger)
		},
	}
}

func pingDB(host, port string) error {
	conn, _ := net.Dial("tcp", host+":"+port)
	_, err := bufio.NewReader(conn).ReadString('\n')

	return err
}
