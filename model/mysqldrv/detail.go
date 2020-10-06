/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"database/sql"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

const qDetail = `SELECT
  response, driver,
  endpoint, content,
  create_at, next_at,
  tried, cur_state
FROM items
WHERE notify_id=? LIMIT 1`

func (d *mysqldrv) Detail(id string) (ret types.Detail, err error) {
	var (
		drv    string
		ep     string
		c      []byte
		create int64
		next   int64
		try    uint32
		state  int
		resp   []byte
	)

	stmt := d.Stmt(qDetail)
	row := stmt.QueryRow(id)
	err = row.Scan(
		&resp,
		&drv,
		&ep,
		&c,
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

	ret = types.Detail{
		Driver:   drv,
		Endpoint: ep,
		Content:  c,
		Response: resp,
		Status: types.Status{
			CreateAt: create,
			NextAt:   next,
			Tried:    try,
			State:    types.State(state),
		},
	}
	return
}
