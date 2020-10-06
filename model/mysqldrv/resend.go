/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import "github.com/raohwork/notify/model"

const qResend = `UPDATE items SET tried=?, cur_state=0 WHERE notify_id=?`

func (d *mysqldrv) Resend(id string, max uint32) (err error) {
	stmt := d.Stmt(qResend)
	res, err := stmt.Exec(id, max-1)
	if err != nil {
		return
	}

	cnt, err := res.RowsAffected()
	if err == nil && cnt != 1 {
		err = &model.E404{}
	}
	return
}
