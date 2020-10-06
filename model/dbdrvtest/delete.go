/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

import (
	"context"
	"testing"
	"time"
)

func (s *suite) testDelete(t *testing.T) {
	sending, sent := make(chan int), make(chan int)
	f := func(ep string, content []byte) (resp []byte, err error) {
		close(sending)
		<-sent
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.send("delete", "ok"); err != nil {
		t.Fatal("cannot create notify: ", err)
	}

	x := func(b bool) {
		if b {
			return
		}
		t.SkipNow()
	}

	<-sending
	x(t.Run("Before", s.testDeleteBefore))
	close(sent)
	time.Sleep(time.Second)
	x(t.Run("After", s.testDeleteAfter))
}

func (s *suite) testDeleteBefore(t *testing.T) {
	if err := s.cl.Delete("delete"); err == nil {
		t.Fatal("deleting running task should be error, but got nothing")
	}

	if _, err := s.cl.Status("delete"); err != nil {
		t.Fatal("cannot get running task: ", err)
	}
}

func (s *suite) testDeleteAfter(t *testing.T) {
	if err := s.cl.Delete("delete"); err != nil {
		t.Fatal("cannot delete task: ", err)
	}

	if _, err := s.cl.Status("delete"); err == nil {
		t.Fatal("should not be able to fetch deleted task info, but no error reported")
	}
}
