/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package httpdrv

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestGetExtract(t *testing.T) {
	d := HTTPGet(http.DefaultClient, nil)
	data := []byte(`{"headers":{"X":["Y"]},"values":{"asd":["qwe"]}}`)

	m, err := d.(*getDrv).extract(data)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if len(m.Headers) != 1 || m.Headers.Get("X") != "Y" {
		t.Errorf("unexpected headers: %+v", m.Headers)
	}

	if len(m.Values) != 1 || m.Values.Get("asd") != "qwe" {
		t.Errorf("unexpected headers: %+v", m.Values)
	}
}

func TestGetSend(t *testing.T) {
	var val url.Values
	h := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		val = r.Form
		w.Write([]byte("OK"))
	}
	srv := httptest.NewServer(http.HandlerFunc(h))

	v := url.Values{}
	v.Set("asd", "qwe")
	data := []byte(`{"values":{"asd": ["qwe"]}}`)
	d := HTTPGet(http.DefaultClient, nil)
	resp, err := d.Send(srv.URL+"/", data)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if s := string(resp); s != "OK" {
		t.Errorf("unexpected response: %s", s)
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("unexpected notify: %+v", val)
	}
}
