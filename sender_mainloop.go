/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"
)

func (w *worker) Start() {
	w.drvStr = make([]string, 0, len(w.drvs))
	for k := range w.drvs {
		w.drvStr = append(w.drvStr, k)
	}

	w.mainloop()
}

func (w *worker) Stop(ctx context.Context) {
	w.cancel()
	for i := uint16(0); i < w.MaxThreads; i++ {
		select {
		case <-ctx.Done():
			return
		case x := <-w.threads:
			close(x.ch)
		}
	}

	ch := make(chan int)
	go func() {
		w.wg.Wait()
		close(ch)
	}()

	select {
	case <-ctx.Done():
	case <-ch:
	}

	close(w.threads)
}

func (w *worker) mainloop() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case t := <-w.threads:
			d, err := w.run(t)
			if err != nil {
				w.threads <- t
				// TODO: log
				break
			}
			t.ch <- d
		}
	}
}

// wait waits a second without blocking w.Stop()
func (w *worker) wait() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case <-w.ctx.Done():
	case <-ctx.Done():
	}
}

func (w *worker) run(t *thread) (d *data, err error) {
	i, err := w.alloc()
	if err != nil {
		return
	}
	if i == nil {
		// nothing to send, wait for a second
		w.job.Unlock()
		w.wait()
		err = errors.New("nothing to send")
		return
	}
	defer w.job.Unlock()

	drv, ok := w.getDrv(i.Driver)
	if !ok {
		// this should never happend
		// TODO: log
		err = fmt.Errorf("got unsupported message: %+v", i)
		return
	}

	d = &data{i: i, drv: drv, ch: w.threads}
	w.job.set(t.id, i.ID)

	return
}
