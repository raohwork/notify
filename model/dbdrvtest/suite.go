/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

import (
	"context"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/raohwork/notify"
	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

const (
	DrvCnt    = 1 // number of notify drivers enabled in test server
	MaxThread = 2 // number of senders in test server
	drvType   = "TEST"
)

// Suite is the integral test suite.
//
// It ensures provided DBDrv is working well together with whole system.
type Suite interface {
	Run(*testing.T)
}

// NewSuite creates a Suite with provided db driver and http server binding address
func NewSuite(drv model.DBDrv, bind string) (ret Suite) {
	arr := strings.Split(bind, ":")
	if len(arr[0]) == 0 {
		arr[0] = "127.0.0.1"
	}
	bind = strings.Join(arr, ":")

	x := &suite{
		dbdrv: drv,
		bind:  bind,
		cl:    types.NewClient("http://"+bind, nil),
	}

	return x
}

type suite struct {
	dbdrv model.DBDrv
	bind  string
	cl    types.Client
}

func (s *suite) waitStart() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	hc := &http.Client{}

	for {
		select {
		case <-ctx.Done():
			log.Fatalf("cannot start api server")
		default:
			_, err := hc.Get("http://" + s.bind)
			if err == nil {
				return
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (s *suite) start(f func(ep string, content []byte) (resp []byte, err error)) (ret notify.APIServer) {
	ret, _ = notify.NewAPI(notify.SenderOptions{
		MaxTries: 3,
		Scheduler: func(driver, notifyID string, lastExec time.Time, tried uint32) (next time.Time, stop bool) {
			next = lastExec.Add(time.Second)
			return
		},
		MaxThreads: MaxThread,
		DBDrv:      s.dbdrv,
	})

	ret.Register(drv(f))
	ret.GetHTTPServer().Addr = s.bind

	go ret.Start()

	s.waitStart()
	return
}

func (s *suite) send(id, uri string) (err error) {
	return s.cl.Send(id, drvType, uri, map[string][]string{})
}

func (s *suite) sendOnce(id, uri string) (err error) {
	return s.cl.SendOnce(id, drvType, uri, map[string][]string{})
}

func (s *suite) Run(t *testing.T) {
	f := func(b bool) {
		if b {
			return
		}
		t.SkipNow()
	}
	f(t.Run("SimpleOK", s.testSimpleOK))
	f(t.Run("SimpleResend", s.testSimpleResend))
	f(t.Run("SimpleDupe", s.testSimpleDupe))
	f(t.Run("SimpleStatus", s.testSimpleStatus))
	f(t.Run("SimpleResult", s.testSimpleResult))
	f(t.Run("SimpleDetail", s.testSimpleDetail))
	f(t.Run("SimpleDelete", s.testSimpleDelete))
	f(t.Run("Clear", s.testClear))
	f(t.Run("Delete", s.testDelete))
}

func (s *suite) waitResult(t time.Duration, ch chan string) (ret string, ok bool) {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	select {
	case <-ctx.Done():
	case ret = <-ch:
		ok = true
	}
	return
}
