/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package smsav8d provides a driver to send SMS using
// https://www.teamplus.tech/product/every8d-value/
package smsav8d

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/raohwork/notify/types"
)

// SMSAV8D is the type of the driver
const SMSAV8D = "SMSAV8D"

// Message defines payload structure
//
// Refer to every8d's API doc for further information
type Message struct {
	// required
	Content string `json:"content"` // MSG field

	// optional
	Subject string `json:"subject"` // SB field
	Time    string `json:"time"`    // ST field
	Retry   int    `json:"retry"`   // RETRYTIME field
}

type drv struct {
	uid string
	pwd string
	hc  *http.Client
}

// New creates a driver to send SMS
//
// The format of payload is defined by Message struct, and endpoint is same as
// "DEST" field in official API doc.
func New(account, password string, hc *http.Client) (ret types.Driver) {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &drv{
		uid: account,
		pwd: password,
		hc:  hc,
	}
}

func (d *drv) Type() (ret string) {
	return SMSAV8D
}

func (d *drv) extract(data []byte) (ret *Message, err error) {
	var m Message

	err = json.Unmarshal(data, &m)
	if err != nil {
		return
	}

	if m.Content == "" {
		err = errors.New("missing required data")
		return
	}

	ret = &m
	return
}

func (d *drv) Verify(data []byte) (err error) {
	_, err = d.extract(data)
	return
}

func (d *drv) Send(ep string, content []byte) (resp []byte, err error) {
	m, err := d.extract(content)
	if err != nil {
		return
	}

	val := url.Values{}
	val.Set("UID", d.uid)
	val.Set("PWD", d.pwd)
	val.Set("MSG", m.Content)
	val.Set("DEST", ep)
	val.Set("SB", m.Subject)
	val.Set("ST", m.Time)
	if m.Retry > 0 {
		val.Set("RETRYTIME", strconv.Itoa(m.Retry))
	}

	uri := "https://oms.every8d.com/API21/HTTP/sendSMS.ashx?" + val.Encode()
	res, err := d.hc.Get(uri)
	if err != nil {
		return
	}
	defer res.Body.Close()
	resp, err = ioutil.ReadAll(res.Body)
	return
}
