/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raohwork/notify/model/dbdrvtest"
)

func TestMysqlDRV(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:test@tcp()/test"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("cannot connect to db: ", err)
	}

	drv, err := New(db, dbdrvtest.DrvCnt, dbdrvtest.MaxThread)
	if err != nil {
		t.Fatal("cannot create mysql db driver: ", err)
	}

	s := dbdrvtest.NewSuite(drv, ":8088")
	s.Run(t)
}
