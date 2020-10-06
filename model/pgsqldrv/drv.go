/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package pgsqldrv provides postgresql db driver
//
// This driver is tested with github.com/jackc/pgx/v4 against postgres 9~13
package pgsqldrv

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/raohwork/notify/model"
)

// I really like to prevent stupid mistakes using language features and software
// architecture:
//
//   1. Using constants to prevent typo.
//   2. Using iota / `qend` constant / array to ensure/check all sql statements are
//      initialized before actually using it.
//   3. Prepare sql statements at first to prevent sql syntax error.
type drv struct {
	*model.DrvBase
	stmts []string
}

// New creates a db driver with postgresql
//
// It creates neccesary table and index if not exists
func New(conn *sql.DB, drvCnt, maxThread int) (ret model.DBDrv, err error) {
	d := &drv{
		DrvBase: model.NewDrvBase(conn),
		stmts:   make([]string, qend),
	}

	const qstr = `CREATE TABLE IF NOT EXISTS items (
notify_id varchar(128) NOT NULL,
driver varchar(16) NOT NULL,
endpoint text NOT NULL,
content bytea NOT NULL,
create_at bigint NOT NULL,
next_at bigint NOT NULL,
tried integer NOT NULL DEFAULT 0,
cur_state smallint NOT NULL DEFAULT 0,
response bytea NULL,
CONSTRAINT items_pk PRIMARY KEY (notify_id)
)`
	const idx1 = `CREATE INDEX IF NOT EXISTS items_pending_idx 
ON items USING btree
(cur_state ASC NULLS LAST)`
	const idx2 = `CREATE INDEX IF NOT EXISTS clear_idx 
ON items USING btree
(create_at ASC NULLS LAST)`

	if _, err = conn.Exec(qstr); err != nil {
		return
	}
	if _, err = conn.Exec(idx1); err != nil {
		return
	}
	if _, err = conn.Exec(idx2); err != nil {
		return
	}

	d.createSql(drvCnt, maxThread)
	if err = d.prepareSql(); err != nil {
		return
	}

	if err == nil {
		ret = d
	}
	return
}

func (d *drv) prepareSql() (err error) {
	for _, qstr := range d.stmts {
		err = d.Prepare(qstr, err)
	}

	return
}

func (d *drv) stmt(key int) (ret *sql.Stmt) {
	return d.Stmt(d.stmts[key])
}

func genvar(begin, cnt int) (ret string) {
	arr := make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		arr = append(arr, "$"+strconv.Itoa(begin+i))
	}

	return strings.Join(arr, ",")
}

const (
	qCreate = iota
	qDelete
	qPending
	qResend
	qResult
	qUpdate
	qClear
	qForceClear
	qStatus
	qDetail
	qend
)

func (d *drv) createSql(drvCnt, maxThread int) {
	d.stmts[qCreate] = `INSERT INTO items
  (notify_id,driver,endpoint,content,create_at,next_at,tried)
VALUES
  ($1,$2,$3,$4,$5,$6,$7)`
	d.stmts[qDelete] = `DELETE FROM items WHERE notify_id=$1`
	d.stmts[qResend] = `UPDATE items SET tried=$1, cur_state=0 WHERE notify_id=$2`
	d.stmts[qResult] = `SELECT response FROM items WHERE notify_id=$1 LIMIT 1`
	d.stmts[qUpdate] = `UPDATE items SET
  tried=$1, next_at=$2, cur_state=$3, response=$4
WHERE notify_id=$5`
	d.stmts[qStatus] = `SELECT create_at, next_at, tried, cur_state FROM items WHERE notify_id=$1`
	d.stmts[qDetail] = `SELECT driver, endpoint, content, response, create_at, next_at, tried, cur_state FROM items WHERE notify_id=$1`

	drvStr := genvar(3, drvCnt)
	curStr := genvar(3+drvCnt, maxThread)
	d.stmts[qPending] = fmt.Sprintf(`SELECT
  notify_id, driver,
  endpoint, content,
  create_at, next_at,
  tried, cur_state
FROM items
WHERE cur_state=0
  AND next_at<=$1
  AND tried<$2
  AND driver IN (%s)
  AND notify_id NOT IN (%s)
ORDER BY next_at ASC
LIMIT 1`, drvStr, curStr)

	curStr = genvar(2, maxThread)
	d.stmts[qClear] = fmt.Sprintf(`DELETE FROM items WHERE create_at < $1 AND cur_state IN (1,2) AND notify_id NOT IN (%s)`, curStr)
	d.stmts[qForceClear] = fmt.Sprintf(`DELETE FROM items WHERE create_at < $1 AND notify_id NOT IN (%s)`, curStr)
}
