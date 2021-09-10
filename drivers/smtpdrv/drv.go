/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package smtpdrv sends notification email using net/smtp
//
// This driver is not mean to be high performance, just provides a simple way to
// send email without applying external services other than your mailbox. To make
// code simpler, it always sends email in MIME+base64 format, slow but always works.
//
// Take a look at TestPayloadAttach(), it also demonstrates how to embed images in
// html email (works in some popular clients including gmail web/app).
package smtpdrv

import (
	"encoding/json"
	"errors"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/raohwork/notify/types"
)

func gen(from mail.Address, addr string, auth smtp.Auth, tlshost string) (ret *drv) {
	ret = &drv{
		from: from,
		conn: &conn{
			addr:    addr,
			auth:    auth,
			tlsHost: tlshost,
			ch:      make(chan *smtp.Client, 1),
		},
	}
	ret.conn.ch <- nil
	return ret
}

// New creates a driver that sends plain text email using net/smtp
//
// This driver does not use endpoint info.
func New(from mail.Address, addr string, auth smtp.Auth, tlshost string) (ret types.Driver) {
	x := gen(from, addr, auth, tlshost)
	x.typ = SMTP
	return x
}

// NewHTML creates a driver that sends html email using net/smtp
//
// This driver does not use endpoint info.
func NewHTML(from mail.Address, addr string, auth smtp.Auth, tlshost string) (ret types.Driver) {
	x := gen(from, addr, auth, tlshost)
	x.typ = SMTPHTML
	return x
}

type drv struct {
	from mail.Address
	conn *conn
	typ  string
}

// SMTP is the driver type that sends plain text email
const SMTP = "SMTP"

// SMTPHTML is the driver type that sends html email
const SMTPHTML = "SMTPHTML"

func (d *drv) Type() (ret string) {
	return d.typ
}

func (d *drv) CheckEP(ep string) (err error) { return }

func (d *drv) Verify(data []byte) (err error) {
	_, err = d.extract(data)
	return
}

func (d *drv) extract(data []byte) (ret *Payload, err error) {
	var x Payload

	err = json.Unmarshal(data, &x)
	if err != nil {
		return
	}

	missing := false
	missing = (missing || x.To == nil)
	missing = (missing || len(x.Content) == 0)
	missing = (missing || x.Subject == "")
	if missing {
		err = errors.New("missing required data")
		return
	}

	ret = &x
	return
}

func genlist(arr []mail.Address) (ret string) {
	lst := make([]string, len(arr))
	for idx, a := range arr {
		lst[idx] = a.String()
	}
	return strings.Join(lst, ", ")
}

func (d *drv) Send(ep string, content []byte) (resp []byte, err error) {
	m, err := d.extract(content)
	if err != nil {
		return
	}

	c, err := d.conn.alloc()
	if err != nil {
		return
	}
	defer d.conn.release(c)

	if err = c.Mail(d.from.String()); err != nil {
		return
	}
	if err = c.Rcpt(m.To.String()); err != nil {
		return
	}
	w, err := c.Data()
	if err != nil {
		return
	}
	buf, err := m.Message(d.from, d.typ == SMTPHTML)
	if err != nil {
		return
	}
	if _, err = w.Write(buf); err != nil {
		return
	}
	if _, err = w.Write([]byte("\r\n.\r\n")); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}

	return
}
