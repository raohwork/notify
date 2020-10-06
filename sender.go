/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

type sender interface {
	Register(types.Driver)
	Start()
	Stop(ctx context.Context)

	drivers() (ret []string)
	driver(typ string) (ret types.Driver, ok bool)
	maxThreads() (ret uint16)
	maxRetry() (ret uint32)
	curID() (ret []string)
}

// SenderOptions defines configurations of internal worker
type SenderOptions struct {
	// how many times to retry before considering as FAILED.
	// which means 1 = do not resend.
	// besides, 0 = math.MaxUint32.
	MaxTries uint32
	// user provided scheduler. nil uses DefaultScheduler
	Scheduler types.Scheduler
	// how many goroutines to do the sending job.
	// 0 will be updated to 1 when creating sender.
	MaxThreads uint16
	// db driver, required
	model.DBDrv
}

func (o *SenderOptions) normalize() (err error) {
	if o.DBDrv == nil {
		return errors.New("empty db driver")
	}

	if o.MaxTries == 0 {
		o.MaxTries = math.MaxUint32
	}
	if o.MaxThreads == 0 {
		o.MaxThreads = 1
	}
	if o.Scheduler == nil {
		o.Scheduler = DefaultScheduler
	}
	return
}

type worker struct {
	SenderOptions
	drvs    map[string]types.Driver
	threads chan *thread
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	drvStr  []string
	job     *jobCtrl
}

// newSender creates a Sender.
func newSender(opt SenderOptions) (ret sender, err error) {
	if err = opt.normalize(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(int(opt.MaxThreads))

	job := newJobCtl(opt.MaxThreads)
	threads := make(chan *thread, opt.MaxThreads)
	for i := uint16(0); i < opt.MaxThreads; i++ {
		x := &thread{
			ch:            make(chan *data),
			SenderOptions: opt,
			wg:            wg,
			id:            i,
			job:           job,
		}

		go x.mainloop()
		threads <- x
	}

	return &worker{
		SenderOptions: opt,
		drvs:          map[string]types.Driver{},
		threads:       threads,
		wg:            wg,
		ctx:           ctx,
		cancel:        cancel,
		job:           job,
	}, nil
}

func (w *worker) Register(drv types.Driver) {
	w.drvs[drv.Type()] = drv
}

func (w *worker) drivers() (ret []string) {
	ret = make([]string, 0, len(w.drvs))
	for k := range w.drvs {
		ret = append(ret, k)
	}
	return
}

func (w *worker) driver(typ string) (ret types.Driver, ok bool) {
	ret, ok = w.drvs[typ]
	return
}

func (w *worker) maxThreads() (ret uint16) {
	return w.MaxThreads
}

func (w *worker) maxRetry() (ret uint32) {
	return w.MaxTries
}

func (w *worker) curID() (ret []string) {
	return w.job.list()
}

func (w *worker) getDrv(typ string) (ret types.Driver, ok bool) {
	ret, ok = w.drvs[typ]
	return
}

func (w *worker) alloc() (ret *model.Item, err error) {
	nowt := time.Now()
	now := nowt.Unix()
	w.job.Lock()
	ret, err = w.Pending(now, w.MaxTries, w.drvStr, w.job.rawList())

	return
}
