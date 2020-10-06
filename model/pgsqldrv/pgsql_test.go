/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package pgsqldrv

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/raohwork/notify/model/dbdrvtest"
)

func TestPgsqlDRV(t *testing.T) {
	dsn := os.Getenv("PGSQL_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@127.0.0.1:5432/test"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("cannot connect to db: ", err)
	}

	drv, err := New(db, dbdrvtest.DrvCnt, dbdrvtest.MaxThread)
	if err != nil {
		t.Fatal("cannot create pgsql db driver: ", err)
	}

	s := dbdrvtest.NewSuite(drv, ":8088")
	s.Run(t)
}
