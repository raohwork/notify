/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"database/sql"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

const qStatus = `SELECT 
  create_at, next_at,
  tried, cur_state
FROM items
WHERE notify_id=? LIMIT 1`

func (d *mysqldrv) Status(id string) (ret types.Status, err error) {
	var (
		create int64
		next   int64
		try    uint32
		state  int
	)

	stmt := d.Stmt(qStatus)
	row := stmt.QueryRow(id)
	err = row.Scan(
		&create,
		&next,
		&try,
		&state,
	)
	if err == sql.ErrNoRows {
		err = &model.E404{}
	}
	if err != nil {
		return
	}

	ret = types.Status{
		CreateAt: create,
		NextAt:   next,
		Tried:    try,
		State:    types.State(state),
	}
	return
}
