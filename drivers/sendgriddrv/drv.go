/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package sendgriddrv sends notification email with sendgrid API
package sendgriddrv

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/raohwork/notify/types"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// New creates a driver that sends notification email via sendgrid
//
// The payload format is json encoded mail.SGMailV3
// https://godoc.org/github.com/sendgrid/sendgrid-go/helpers/mail#SGMailV3
//
// This driver does not use endpoint info.
func New(key string, cl *http.Client) (ret types.Driver) {
	if cl == nil {
		cl = http.DefaultClient
	}
	return &drv{
		key: key,
		hc:  cl,
	}
}

type drv struct {
	key string
	hc  *http.Client
}

// Sendgrid is driver type of New()
const Sendgrid = "SENDGRID"

func (d *drv) Type() (ret string) {
	return Sendgrid
}

func (d *drv) CheckEP(ep string) (err error) { return }

func (d *drv) Verify(data []byte) (err error) {
	_, err = d.extract(data)
	return
}

func (d *drv) extract(data []byte) (ret *mail.SGMailV3, err error) {
	var x mail.SGMailV3

	err = json.Unmarshal(data, &x)
	if err != nil {
		return
	}

	missing := false
	missing = missing || x.From == nil
	missing = missing || len(x.Content) == 0
	if missing {
		err = errors.New("missing required data")
		return
	}

	ret = &x
	return
}

func (d *drv) Send(ep string, content []byte) (resp []byte, err error) {
	m, err := d.extract(content)
	if err != nil {
		return
	}

	cl := sendgrid.NewSendClient(d.key)
	sgres, err := cl.Send(m)
	resp, _ = json.Marshal(sgres)

	return
}
