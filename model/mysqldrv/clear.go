/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import (
	"time"
)

const qClear = "DELETE FROM items WHERE create_at < ? AND cur_state IN (1,2) AND notify_id NOT IN (%s)"

var qClearReal string

func (d *mysqldrv) Clear(t time.Time, cur []string) (err error) {
	stmt := d.Stmt(qClearReal)
	args := make([]interface{}, 1, len(cur)+1)
	args[0] = t.Unix()
	for _, id := range cur {
		args = append(args, id)
	}

	_, err = stmt.Exec(args...)
	return
}

const qForceClear = "DELETE FROM items WHERE create_at < ? AND notify_id NOT IN (%s)"

var qForceClearReal string

func (d *mysqldrv) ForceClear(t time.Time, cur []string) (err error) {
	stmt := d.Stmt(qForceClearReal)
	args := make([]interface{}, 1, len(cur)+1)
	args[0] = t.Unix()
	for _, id := range cur {
		args = append(args, id)
	}
	_, err = stmt.Exec(args...)
	return
}
