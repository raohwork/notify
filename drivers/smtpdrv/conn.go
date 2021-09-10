/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package smtpdrv

import (
	"crypto/tls"
	"net/smtp"
)

type conn struct {
	ch      chan *smtp.Client
	addr    string
	auth    smtp.Auth
	tlsHost string
}

func (c *conn) release(cl *smtp.Client) {
	c.ch <- cl
}

func (c *conn) alloc() (ret *smtp.Client, err error) {
	x := <-c.ch
	if x == nil {
		return c.dial()
	}
	if err = x.Noop(); err != nil {
		x.Close()
		return c.dial()
	}

	ret = x
	return
}

func (c *conn) dial() (ret *smtp.Client, err error) {
	x, err := smtp.Dial(c.addr)
	if err != nil {
		return
	}
	if c.tlsHost != "" {
		err = x.StartTLS(&tls.Config{ServerName: c.tlsHost})
		if err != nil {
			return
		}
	}
	if err = x.Auth(c.auth); err != nil {
		return
	}

	ret = x
	return
}
