/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

import (
	"net/http"
	"time"

	"github.com/raohwork/notify"
	"github.com/raohwork/notify/drivers/httpdrv"
	"github.com/raohwork/notify/model"
)

type srv struct {
	notify.APIServer
}

func newSrv(drv model.DBDrv) (ret *srv, err error) {
	opt := notify.SenderOptions{
		MaxTries: 3,
		Scheduler: func(driver, notifyID string, lastExec time.Time, tried uint32) (next time.Time, stop bool) {
			next = lastExec.Add(5 * time.Second)
			return
		},
		MaxThreads: 2,
		DBDrv:      drv,
	}

	api, err := notify.NewAPI(opt)
	if err != nil {
		return
	}
	api.Register(httpdrv.HTTPGet(&http.Client{}, nil))

	ret = &srv{APIServer: api}
	return
}
