/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/raohwork/notify/types"
)

func (s *suite) testSimpleOK(t *testing.T) {
	const uri = "ok"
	ch := make(chan string)
	f := func(ep string, content []byte) (resp []byte, err error) {
		go func() { ch <- ep }()
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.send("simple", uri); err != nil {
		t.Fatal("cannot create notify: ", err)
	}

	res, ok := s.waitResult(5*time.Second, ch)
	if !ok {
		t.Fatal("notify is not sent in 5 seconds")
	}

	if res != uri {
		t.Log("expect: ok")
		t.Log("actual:", res)
		t.Fatal("expected result")
	}
}

func (s *suite) testSimpleResend(t *testing.T) {
	const uri = "ok"
	ch := make(chan string)
	f := func(ep string, content []byte) (resp []byte, err error) {
		go func() { ch <- ep }()
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.cl.Resend("simple"); err != nil {
		t.Fatal("cannot create notify: ", err)
	}

	res, ok := s.waitResult(5*time.Second, ch)
	if !ok {
		t.Fatal("notify is not sent in 5 seconds")
	}

	if res != uri {
		t.Log("expect: ok")
		t.Log("actual:", res)
		t.Fatal("expected result")
	}
}

func (s *suite) testSimpleDupe(t *testing.T) {
	f := func(ep string, content []byte) (resp []byte, err error) {
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.send("simple", "ok"); err == nil {
		t.Fatal("sending duplicated id should return error, but got none")
	}
}

func (s *suite) testSimpleStatus(t *testing.T) {
	f := func(ep string, content []byte) (resp []byte, err error) {
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	x, err := s.cl.Status("simple")
	if err != nil {
		t.Fatal("cannot send status request: ", err)
	}

	if x.State != types.SUCCESS {
		t.Logf("expect: %d", types.SUCCESS)
		t.Logf("actual: %d", x.State)
		t.Fatal("unexpected state")
	}
}

func (s *suite) testSimpleResult(t *testing.T) {
	f := func(ep string, content []byte) (resp []byte, err error) {
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	buf, err := s.cl.Result("simple")
	if err != nil {
		t.Fatal("cannot send result request: ", err)
	}

	if x := string(buf); x != "ok" {
		t.Log("expect: ok")
		t.Log("actual:", string(buf))
		t.Fatal("unexpected result")
	}
}

func (s *suite) testSimpleDetail(t *testing.T) {
	f := func(ep string, content []byte) (resp []byte, err error) {
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	st, err := s.cl.Status("simple")
	if err != nil {
		t.Fatal("cannot get status: ", err)
	}

	x, err := s.cl.Detail("simple")
	if err != nil {
		t.Fatal("cannot get detail: ", err)
	}

	if !reflect.DeepEqual(st, x.Status) {
		t.Logf("expect: %+v", st)
		t.Logf("actual: %+v", x.Status)
		t.Error("unexpected status")
	}
	if str := string(x.Content); str != "{}" {
		t.Logf("expect: {}")
		t.Logf("actual: %s", str)
		t.Error("unexpected content")
	}
	if str := string(x.Response); str != "ok" {
		t.Logf("expect: ok")
		t.Logf("actual: %s", str)
		t.Error("unexpected response")
	}
	if x.Driver != drvType {
		t.Logf("expect: %s", drvType)
		t.Logf("actual: %s", x.Driver)
		t.Error("unexpected driver")
	}
	if x.Endpoint != "ok" {
		t.Logf("expect: ok")
		t.Logf("actual: %s", x.Endpoint)
		t.Error("unexpected endpoint")
	}
	return
}

func (s *suite) testSimpleDelete(t *testing.T) {
	f := func(ep string, content []byte) (resp []byte, err error) {
		return []byte(ep), nil
	}
	api := s.start(f)
	defer api.Shutdown(context.Background())

	if err := s.cl.Delete("simple"); err != nil {
		t.Fatal("cannot delet notify: ", err)
	}

	detail, err := s.cl.Detail("simple")
	if err == nil {
		err = fmt.Errorf("got item:\n%+v", detail)
	} else {
		err = nil
	}

	return
}
