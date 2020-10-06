/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package httpdrv

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/raohwork/notify/types"
)

func FormPostMsg(val url.Values) (ret *PostMsg) {
	ret = &PostMsg{
		Body: []byte(val.Encode()),
	}
	ret.Headers.Set("Content-Type", "application/x-www-form-urlencoded")

	return
}

func JSONPostMsg(i interface{}) (ret *PostMsg, err error) {
	buf, err := json.Marshal(i)
	if err != nil {
		return
	}
	ret = &PostMsg{
		Body: buf,
	}
	ret.Headers.Set("Content-Type", "application/json")

	return
}

// PostMsg defines driver specific parameters for HTTPPost()
type PostMsg struct {
	// additional header to send
	Headers http.Header `json:"headers"`
	// post body
	Body []byte `json:"body"`
}

// POST is driver type of HTTPPost()
const POST = "HTTPPOST"

type postDrv struct {
	cl *http.Client
	v  Validator
}

// HTTPPost creates a driver that delivers notification via HTTP POST
func HTTPPost(cl *http.Client, v Validator) (ret types.Driver) {
	if v == nil {
		v = DefaultValidator
	}
	return &postDrv{
		cl: cl,
		v:  v,
	}
}

func (d *postDrv) Type() (ret string) {
	return POST
}

func (d *postDrv) Verify(buf []byte) (err error) {
	_, err = d.extract(buf)
	return
}

func (d *postDrv) extract(buf []byte) (ret *PostMsg, err error) {
	var x PostMsg

	err = json.Unmarshal(buf, &x)
	if err != nil {
		return
	}

	ret = &x
	return
}

func (d *postDrv) Send(ep string, data []byte) (resp []byte, err error) {
	msg, err := d.extract(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", ep, bytes.NewReader(msg.Body))
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
