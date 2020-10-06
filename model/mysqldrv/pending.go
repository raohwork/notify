/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"database/sql"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

const qPending = `SELECT
  notify_id, driver,
  endpoint, content,
  create_at, next_at,
  tried, cur_state
FROM items
WHERE cur_state=0
  AND next_at<=?
  AND tried<?
  AND driver IN (%s)
  AND notify_id NOT IN (%s)
ORDER BY next_at ASC
LIMIT 1`

var qPendingReal string

func (d *mysqldrv) Pending(now int64, max uint32, drvs, ids []string) (ret *model.Item, err error) {
	var (
		id     string
		drv    string
		ep     string
		c      []byte
		create int64
		next   int64
		try    uint32
		state  int
	)

	params := make([]interface{}, 0, len(drvs)+len(ids)+1)
	params = append(params, now, max)
	for _, d := range drvs {
		params = append(params, d)
	}
	for _, d := range ids {
		params = append(params, d)
	}

	if err != nil {
		return
	}

	stmt := d.Stmt(qPendingReal)
	row := stmt.QueryRow(params...)
	err = row.Scan(
		&id,
		&drv,
		&ep,
		&c,
		&create,
		&next,
		&try,
		&state,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}

	ret = &model.Item{
		ID:       id,
		Driver:   drv,
		Endpoint: ep,
		Content:  c,
		CreateAt: create,
		NextAt:   next,
		Tried:    try,
		State:    types.State(state),
	}
	return
}
