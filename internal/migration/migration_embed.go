//go:build migration
// +build migration

package migration

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/scripts/migration"
	"github.com/sirupsen/logrus"
)

func Run() {
	if !conf.CfgIf("Migration") {
		logrus.Infoln("skip migrate because not add Migration feature in config.yaml")
		return
	}

	var (
		db        *sql.DB
		dbName    string
		dbDriver  database.Driver
		srcDriver source.Driver
		err, err2 error
	)

	if conf.CfgIf("MySQL") {
		dbName = conf.MysqlSetting.DBName
		db, err = sql.Open("mysql", conf.MysqlSetting.Dsn()+"&multiStatements=true")
	} else if conf.CfgIf("PostgreSQL") || conf.CfgIf("Postgres") {
		dbName = (*conf.PostgresSetting)["DBName"]
		db, err = sql.Open("postgres", conf.PostgresSetting.Dsn())
	} else if conf.CfgIf("Sqlite3") {
		db, err = sql.Open("sqlite3", conf.Sqlite3Setting.Dsn())
	} else {
		dbName = conf.MysqlSetting.DBName
		db, err = sql.Open("mysql", conf.MysqlSetting.Dsn())
	}
	if err != nil {
		logrus.Errorf("initial db for migration failed: %s", err)
		return
	}

	if conf.CfgIf("MySQL") {
		srcDriver, err = iofs.New(migration.Files, "mysql")
		dbDriver, err2 = mysql.WithInstance(db, &mysql.Config{})
	} else if conf.CfgIf("PostgreSQL") || conf.CfgIf("Postgres") {
		srcDriver, err = iofs.New(migration.Files, "postgres")
		dbDriver, err2 = postgres.WithInstance(db, &postgres.Config{})
	} else if conf.CfgIf("Sqlite3") {
		srcDriver, err = iofs.New(migration.Files, "sqlite3")
		dbDriver, err2 = sqlite3.WithInstance(db, &sqlite3.Config{})
	} else {
		srcDriver, err = iofs.New(migration.Files, "mysql")
		dbDriver, err2 = mysql.WithInstance(db, &mysql.Config{})
	}

	if err2 != nil {
		logrus.Errorf("new database driver failed: %s", err)
		return
	} else {
		defer dbDriver.Close()
	}
	if err != nil {
		logrus.Errorf("new source driver failed: %s", err)
		return
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, dbName, dbDriver)
	if err != nil {
		logrus.Errorf("new migrate instance failed: %s", err)
		return
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		logrus.Errorf("migrate up failed: %s", err)
		return
	}
	logrus.Infoln("migrate up success")
}
