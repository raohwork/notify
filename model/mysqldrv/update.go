/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import "github.com/raohwork/notify/types"

const qUpdate = `UPDATE items SET
  tried=?, next_at=?, cur_state=?, response=?
WHERE notify_id=?`

func (d *mysqldrv) Update(id string, tried uint32, next int64, state types.State, resp []byte) (err error) {
	stmt := d.Stmt(qUpdate)
	_, err = stmt.Exec(tried, next, state, resp, id)
	return
}
