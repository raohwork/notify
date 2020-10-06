/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"sync"
	"time"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

type data struct {
	i   *model.Item
	drv types.Driver
	ch  chan *thread
}

type thread struct {
	ch chan *data
	SenderOptions
	wg  *sync.WaitGroup
	id  uint16
	job *jobCtrl
}

func (t *thread) mainloop() {
	for d := range t.ch {
		t.run(d.i, d.drv)
		d.ch <- t
	}

	t.wg.Done()
}

func (t *thread) run(i *model.Item, drv types.Driver) {
	defer func() {
		t.job.lset(t.id, "")
	}()

	now := time.Now()
	next, stop := t.Scheduler(drv.Type(), i.ID, now, i.Tried)
	state := types.PENDING
	i.Tried++
	i.NextAt = next.Unix()
	if stop {
		i.Tried = t.MaxTries
	}
	if i.Tried >= t.MaxTries {
		state = types.FAILED
	}

	resp, err := drv.Send(i.Endpoint, i.Content)

	if err == nil {
		state = types.SUCCESS
	} else {
		// TODO: log error
		if len(resp) == 0 {
			resp = []byte(err.Error())
		}
	}

	t.Update(i.ID, i.Tried, i.NextAt, state, resp)
}
