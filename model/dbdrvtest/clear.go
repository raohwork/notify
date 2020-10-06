/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raohwork/notify/types"
)

func (s *suite) testClear(t *testing.T) {
	ok, failed := make(chan int), make(chan int, 3)
	defer close(failed)
	f := func(ep string, content []byte) (resp []byte, err error) {
		if ep == "ok" {
			close(ok)
			return []byte(ep), nil
		}
		failed <- 0
		return []byte(ep), errors.New("err")
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.send("simple1", "ok"); err != nil {
		t.Fatal("cannot create notify1: ", err)
	}
	if err := s.send("simple2", "fail"); err != nil {
		t.Fatal("cannot create notify2: ", err)
	}

	x := func(b bool) {
		if b {
			return
		}
		t.SkipNow()
	}

	<-ok // wait for success notify
	time.Sleep(time.Second)
	x(t.Run("Pending", s.testClearPending))
	x(t.Run("Success", s.testClearSuccess))
	<-failed // wait for failed notify
	<-failed
	<-failed
	time.Sleep(time.Second)
	x(t.Run("Failed", s.testClearFailed))
}

func (s *suite) testClearPending(t *testing.T) {
	st, err := s.cl.Status("simple2")
	if err != nil {
		t.Fatal("cannot get status: ", err)
	}
	if st.State != types.PENDING {
		t.Fatal("unexpected state: ", st.State)
	}

	d := time.Now()
	if err = s.cl.Clear(d); err != nil {
		t.Fatal("cannot call clear: ", err)
	}

	st, err = s.cl.Status("simple2")
	if err != nil {
		t.Fatal("cannot get status: ", err)
	}
	if st.State != types.PENDING {
		t.Fatal("unexpected state: ", st.State)
	}
}

func (s *suite) testClearSuccess(t *testing.T) {
	st, err := s.cl.Status("simple1")
	if err == nil {
		t.Logf("%+v", st)
		t.Fatal("simple1 should be cleared, but still there")
	}
}

func (s *suite) testClearFailed(t *testing.T) {
	st, err := s.cl.Status("simple2")
	if err != nil {
		t.Logf("%+v", st)
		t.Fatal("cannot get status: ", err)
	}

	d := time.Now()
	if err = s.cl.Clear(d); err != nil {
		t.Fatal("cannot call clear: ", err)
	}

	st, err = s.cl.Status("simple2")
	if err == nil {
		t.Logf("%+v", st)
		t.Fatal("simple1 should be cleared, but still there")
	}
}
