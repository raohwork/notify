/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package mysqldrv

import "github.com/raohwork/notify/model"

const qCreate = `INSERT INTO items
  (notify_id,driver,endpoint,content,create_at,next_at,tried)
VALUES
  (?,?,?,?,?,?,?)`

func (d *mysqldrv) Create(i *model.Item) (err error) {
	stmt := d.Stmt(qCreate)
	_, err = stmt.Exec(
		i.ID, i.Driver,
		i.Endpoint, i.Content,
		i.CreateAt, i.NextAt, i.Tried,
	)
	return
}
