/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package httpdrv

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/raohwork/notify/types"
)

// GetMsg defines driver specific parameters for HTTPGet()
type GetMsg struct {
	// additional headers to send
	Headers http.Header `json:"headers,omitempty"`
	// url query string to send. the value is encoded and appended to endpoint
	Values url.Values `json:"values,omitempty"`
}

// GET is driver type of HTTPGet()
const GET = "HTTPGET"

type getDrv struct {
	cl *http.Client
	v  Validator
}

// HTTPGet creates a driver that delivers notification via HTTP GET
func HTTPGet(cl *http.Client, v Validator) (ret types.Driver) {
	if v == nil {
		v = DefaultValidator
	}
	return &getDrv{
		cl: cl,
		v:  v,
	}
}

func (d *getDrv) CheckEP(ep string) (err error) {
	u, err := url.Parse(ep)
	if err != nil {
		return
	}

	if len(u.Query()) > 0 {
		err = errors.New("there cannot be any queries in url")
	}
	return
}

func (d *getDrv) Type() (ret string) {
	return GET
}

func (d *getDrv) Verify(buf []byte) (err error) {
	_, err = d.extract(buf)
	return
}

func (d *getDrv) extract(buf []byte) (ret *GetMsg, err error) {
	var x GetMsg

	err = json.Unmarshal(buf, &x)
	if err != nil {
		return
	}

	ret = &x
	return
}

func (d *getDrv) Send(ep string, data []byte) (resp []byte, err error) {
	msg, err := d.extract(data)
	if err != nil {
		return
	}

	if len(msg.Values) > 0 {
		ep += "?" + msg.Values.Encode()
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return
	}
	req.Header = msg.Headers

	res, err := d.cl.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	resp, _ = ioutil.ReadAll(res.Body)

	err = d.v(res.StatusCode, res.Header, resp)
	return
}
