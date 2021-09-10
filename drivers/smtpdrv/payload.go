/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package smtpdrv

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/mail"
	"net/textproto"
)

func quote(s string) (ret string) {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

type Attach struct {
	Type       string
	RawContent string
	ID         string
	Filename   string
}

func (a Attach) write(w *multipart.Writer, attach bool) (err error) {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", a.Type)
	h.Set("Content-Transfer-Encoding", "base64")
	if attach {
		att := "attachment"
		if a.Filename != "" {
			var fn string
			fn = quote(a.Filename)
			att += "; filename=" + fn
		}
		h.Set("Content-Disposition", att)
	}
	if a.ID != "" {
		h.Set("Content-ID", "<"+a.ID+">")
	}
	x, err := w.CreatePart(h)
	if err != nil {
		return
	}
	x.Write([]byte(base64.StdEncoding.EncodeToString([]byte(a.RawContent))))
	return
}

type Payload struct {
	To      *mail.Address
	CC      []mail.Address
	BCC     []mail.Address
	Subject string
	Content string
	Attach  []Attach
}

func (p *Payload) Message(from mail.Address, html bool) (ret []byte, err error) {
	return p.message(from, html, "")
}

func (p *Payload) message(from mail.Address, html bool, bound string) (ret []byte, err error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	if bound != "" {
		err = w.SetBoundary(bound)
		if err != nil {
			return
		}
	}
	wbuf := func(err error, str string) error {
		if err != nil {
			return err
		}
		_, err = buf.WriteString(str + "\r\n")
		return err
	}
	err = wbuf(err, "MIME-Version: 1.0")
	err = wbuf(err, "Content-Type: multipart/mixed; boundary="+w.Boundary())
	err = wbuf(err, "From: "+from.String())
	err = wbuf(err, "To: "+p.To.String())
	err = wbuf(err, "Subject: "+quote(p.Subject))
	if len(p.CC) > 0 {
		err = wbuf(err, "Cc: "+genlist(p.CC))
	}
	if len(p.BCC) > 0 {
		err = wbuf(err, "Bcc: "+genlist(p.BCC))
	}
	err = wbuf(err, "")
	if err != nil {
		return
	}

	// body
	body := Attach{
		Type:       "text/plain",
		RawContent: p.Content,
	}
	if html {
		body.Type = "text/html; charset=UTF-8"
	}
	if err = body.write(w, false); err != nil {
		return
	}

	for _, a := range p.Attach {
		if err = a.write(w, true); err != nil {
			return
		}
	}
	if err = w.Close(); err != nil {
		return
	}

	ret = buf.Bytes()
	return
}
