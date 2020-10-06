/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import "errors"

const qDelete = "DELETE FROM items WHERE notify_id=?"

func (d *mysqldrv) Delete(id string, ids []string) (err error) {
	for _, i := range ids {
		if id == i {
			return errors.New("notification is processing, cannot delete")
		}
	}
	stmt := d.Stmt(qDelete)
	_, err = stmt.Exec(id)
	return
}
