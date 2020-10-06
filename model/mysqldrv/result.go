/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"database/sql"

	"github.com/raohwork/notify/model"
)

const qResult = `SELECT response FROM items WHERE notify_id=? LIMIT 1`

func (d *mysqldrv) Result(id string) (ret []byte, err error) {
	stmt := d.Stmt(qResult)
	row := stmt.QueryRow(id)
	err = row.Scan(&ret)
	if err == sql.ErrNoRows {
		err = &model.E404{}
	}
	return
}
