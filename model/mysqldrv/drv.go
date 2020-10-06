/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package mysqldrv provides mysql db driver
//
// This driver is tested with github.com/go-sql-driver/mysql against mysql 5.7/8
package mysqldrv

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/raohwork/notify/model"
)

type mysqldrv struct {
	*model.DrvBase
}

// New creates a db driver with mysql
//
// It will create neccessary table is not exists.
func New(conn *sql.DB, drvCnt int, maxThread int) (ret model.DBDrv, err error) {
	d := &mysqldrv{
		DrvBase: model.NewDrvBase(conn),
	}

	// create table if not exists
	if err = d.table(); err != nil {
		return
	}

	err = d.Prepare(qCreate, err)
	err = d.Prepare(qResend, err)
	err = d.Prepare(qUpdate, err)
	err = d.Prepare(qResult, err)
	err = d.Prepare(qDelete, err)
	err = d.Prepare(qStatus, err)
	err = d.Prepare(qDetail, err)
	drv := strings.Repeat(",?", drvCnt)[1:]
	ids := strings.Repeat(",?", maxThread)[1:]
	qPendingReal = fmt.Sprintf(qPending, drv, ids)
	err = d.Prepare(qPendingReal, err)
	qClearReal = fmt.Sprintf(qClear, ids)
	err = d.Prepare(qClearReal, err)
	qForceClearReal = fmt.Sprintf(qForceClear, ids)
	err = d.Prepare(qForceClearReal, err)

	if err == nil {
		ret = d
	}
	return
}

const qTable = "CREATE TABLE IF NOT EXISTS items (`notify_id` varchar(128) NOT NULL PRIMARY KEY, `driver` varchar(16) NOT NULL, `endpoint` text NOT NULL, `content` blob NOT NULL, `create_at` bigint NOT NULL, `next_at` bigint NOT NULL, `tried` int UNSIGNED NOT NULL DEFAULT 0, `cur_state` tinyint(1) NOT NULL DEFAULT 0, `response` blob NULL, INDEX `pending_key` (`next_at`), INDEX `creation_key` (`create_at`))"

func (d *mysqldrv) table() (err error) {
	_, err = d.DB.Exec(qTable)
	return
}
