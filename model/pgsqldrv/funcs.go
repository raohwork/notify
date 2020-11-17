/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package pgsqldrv

import (
	"database/sql"
	"errors"
	"time"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

func (d *drv) Create(i *model.Item) (err error) {
	stmt := d.stmt(qCreate)
	_, err = stmt.Exec(
		i.ID, i.Driver,
		i.Endpoint, i.Content,
		i.CreateAt, i.NextAt, i.Tried,
	)
	return
}

func (d *drv) Delete(id string, ids []string) (err error) {
	for _, i := range ids {
		if id == i {
			return errors.New("notification is processing, cannot delete")
		}
	}
	stmt := d.stmt(qDelete)
	_, err = stmt.Exec(id)
	return
}

func (d *drv) Resend(id string, max uint32) (err error) {
	stmt := d.stmt(qResend)
	_, err = stmt.Exec(max-1, id)
	return
}

func (d *drv) Result(id string) (ret []byte, err error) {
	stmt := d.stmt(qResult)
	row := stmt.QueryRow(id)
	err = row.Scan(&ret)
	return
}

func (d *drv) Update(id string, tried uint32, next int64, state types.State, resp []byte) (err error) {
	stmt := d.stmt(qUpdate)
	_, err = stmt.Exec(tried, next, state, resp, id)
	return
}

func (d *drv) Clear(t time.Time, cur []string) (err error) {
	stmt := d.stmt(qClear)
	args := make([]interface{}, 1, len(cur)+1)
	args[0] = t.Unix()
	for _, id := range cur {
		args = append(args, id)
	}
	_, err = stmt.Exec(args...)
	return
}

func (d *drv) ForceClear(t time.Time, cur []string) (err error) {
	stmt := d.stmt(qForceClear)
	args := make([]interface{}, 1, len(cur)+1)
	args[0] = t.Unix()
	for _, id := range cur {
		args = append(args, id)
	}
	_, err = stmt.Exec(args...)
	return
}

func (d *drv) Status(id string) (ret types.Status, err error) {
	var (
		create int64
		next   int64
		try    uint32
		state  int
	)

	stmt := d.stmt(qStatus)
	row := stmt.QueryRow(id)
	err = row.Scan(
		&create,
		&next,
		&try,
		&state,
	)
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

func (d *drv) Detail(id string) (ret types.Detail, err error) {
	var (
		drv    string
		ep     string
		c      []byte
		resp   []byte
		create int64
		next   int64
		try    uint32
		state  int
	)

	stmt := d.stmt(qDetail)
	row := stmt.QueryRow(id)
	err = row.Scan(
		&drv,
		&ep,
		&c,
		&resp,
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

func (d *drv) Pending(now int64, max uint32, drvs, ids []string) (ret *model.Item, err error) {
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

	stmt := d.stmt(qPending)
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
