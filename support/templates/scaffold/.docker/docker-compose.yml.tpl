version: "3.8"

services:{{if eq .dbAdapter "mysql"}}
  mysql-master:
    image: bitnami/mysql:8.0
    ports:
      - "23306:3306"
    environment:
      - MYSQL_REPLICATION_MODE=master
      - MYSQL_REPLICATION_USER=replicator
      - MYSQL_REPLICATION_PASSWORD=whatever
      - MYSQL_ROOT_PASSWORD=whatever
      - MYSQL_USER=mysql
      - MYSQL_PASSWORD=whatever

  mysql-slave:
    image: bitnami/mysql:8.0
    ports:
      - "23307:3306"
    depends_on:
      - mysql-master
    environment:
      - MYSQL_REPLICATION_MODE=slave
      - MYSQL_REPLICATION_USER=replicator
      - MYSQL_REPLICATION_PASSWORD=whatever
      - MYSQL_MASTER_HOST=mysql-master
      - MYSQL_MASTER_PORT_NUMBER=3306
      - MYSQL_MASTER_ROOT_PASSWORD=whatever{{else if eq .dbAdapter "postgres"}}
  psql-master:
    image: bitnami/postgresql:12
    ports:
      - "25432:5432"
    environment:
      - POSTGRESQL_REPLICATION_MODE=master
      - POSTGRESQL_REPLICATION_USER=replicator
      - POSTGRESQL_REPLICATION_PASSWORD=whatever
      - POSTGRESQL_USERNAME=postgres
      - POSTGRESQL_PASSWORD=whatever

  psql-slave:
    image: bitnami/postgresql:12
    ports:
      - "25433:5432"
    depends_on:
      - psql-master
    environment:
      - POSTGRESQL_REPLICATION_MODE=slave
      - POSTGRESQL_REPLICATION_USER=replicator
      - POSTGRESQL_REPLICATION_PASSWORD=whatever
      - POSTGRESQL_MASTER_HOST=psql-master
      - POSTGRESQL_PASSWORD=whatever
      - POSTGRESQL_MASTER_PORT_NUMBER=5432{{end}}

  redis-standalone:
    image: bitnami/redis:5.0.5
    ports:
      - "26379:6379"
    environment:
      - ALLOW_EMPTY_PASSWORD=yes

  redis-sentinel:
    image: bitnami/redis-sentinel:5.0.5
    ports:
      - "26380:26379"
    environment:
      - REDIS_MASTER_HOST=redis-master
      - REDIS_MASTER_SET=mymaster
    depends_on:
      - redis-master
      - redis-slave

  redis-master:
    image: bitnami/redis:5.0.5
    ports:
      - "26381:6379"
    environment:
      - ALLOW_EMPTY_PASSWORD=yes
      - REDIS_REPLICATION_MODE=master

  redis-slave:
    image: bitnami/redis:5.0.5
    ports:
      - "26382:6379"
    environment:
      - ALLOW_EMPTY_PASSWORD=yes
      - REDIS_REPLICATION_MODE=slave
      - REDIS_MASTER_HOST=redis-master
