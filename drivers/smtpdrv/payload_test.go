/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package smtpdrv

import (
	"bytes"
	"net/mail"
	"strings"
	"testing"
)

func TestPayloadSimple(t *testing.T) {
	expect := strings.Join([]string{
		"MIME-Version: 1.0",
		"Content-Type: multipart/mixed; boundary=fuck",
		"From: =?utf-8?q?=E4=B8=AD=E6=96=87=E5=90=8D?= <wtf@example.com>",
		"To: \"asc ii\" <asc@i.i>",
		"Subject: =?UTF-8?B?5Lit5paH5qiZ6aGM?=",
		"",
		"--fuck",
		"Content-Transfer-Encoding: base64",
		"Content-Type: text/plain",
		"",
		"5Lit5paH5YWn5a65",
		"--fuck--",
		"",
	}, "\r\n")
	p := &Payload{
		To: &mail.Address{
			Name:    "asc ii",
			Address: "asc@i.i",
		},
		Subject: "中文標題",
		Content: "中文內容",
	}

	actual, err := p.message(mail.Address{
		Name:    "中文名",
		Address: "wtf@example.com",
	}, false, "fuck")
	if err != nil {
		t.Fatal(err)
	}

	if s := string(actual); s != expect {
		arr := bytes.Split([]byte(expect), []byte("\r\n"))
		for _, a := range arr {
			t.Log(a)
		}
		t.Log()
		t.Log()
		arr = bytes.Split(actual, []byte("\r\n"))
		for _, a := range arr {
			t.Log(a)
		}
		t.Fatal(s)
	}
}

func TestPayloadAttach(t *testing.T) {
	expect := strings.Join([]string{
		"MIME-Version: 1.0",
		"Content-Type: multipart/mixed; boundary=fuck",
		"From: =?utf-8?q?=E4=B8=AD=E6=96=87=E5=90=8D?= <service@shoutloud.work>",
		"To: \"asc ii\" <ronmi.ren@gmail.com>",
		"Subject: =?UTF-8?B?MTIz5ris6Kmm?=",
		"",
		"--fuck",
		"Content-Transfer-Encoding: base64",
		"Content-Type: text/plain",
		"",
		"5ris6Kmm6YO15Lu2",
		"--fuck",
		"Content-Disposition: attachment; filename==?UTF-8?B?YXNkLmpwZw==?=",
		"Content-Id: <2.jpg>",
		"Content-Transfer-Encoding: base64",
		"Content-Type: image/jpeg",
		"",
		"WVdKalpHVm0=",
		"--fuck",
		"Content-Disposition: attachment; filename==?UTF-8?B?YXNkLmpwZw==?=",
		"Content-Id: <1.jpg>",
		"Content-Transfer-Encoding: base64",
		"Content-Type: image/jpeg",
		"",
		"WVdKalpHVm0=",
		"--fuck",
		"Content-Disposition: attachment",
		"Content-Id: <nofn>",
		"Content-Transfer-Encoding: base64",
		"Content-Type: image/jpeg",
		"",
		"WVdKalpHVm0=",
		"--fuck--",
		"",
	}, "\r\n")
	pic := `YWJjZGVm`
	from := mail.Address{
		Name:    "中文名",
		Address: "service@shoutloud.work",
	}
	to := mail.Address{
		Name:    "asc ii",
		Address: "ronmi.ren@gmail.com",
	}
	p := &Payload{
		To:      &to,
		Subject: "123測試",
		Content: `<!doctype html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width" />
<title>測試123</title>
</head>
<body>
<h3>測試</h3>
<div>
<img src="cid:nofn" alt="測試影像" />
</div>
<div>
<img src="cid:1.jpg" alt="測試影像" />
</div>
<div>
<img src="cid:2.jpg" alt="測試影像" />
</div>
</html>`,
		Attach: []Attach{
			{
				Type:       "image/jpeg",
				RawContent: pic,
				Filename:   "asd.jpg",
				ID:         "2.jpg",
			},
			{
				Type:       "image/jpeg",
				RawContent: pic,
				Filename:   "asd.jpg",
				ID:         "1.jpg",
			},
			{
				Type:       "image/jpeg",
				RawContent: pic,
				ID:         "nofn",
			},
		},
	}

	actual, err := p.message(from, true, "fuck")
	if err != nil {
		t.Fatal(err)
	}

	if s := string(actual); s != expect {
		arr := bytes.Split([]byte(expect), []byte("\r\n"))
		for _, a := range arr {
			t.Log(a)
		}
		t.Log()
		t.Log()
		arr = bytes.Split(actual, []byte("\r\n"))
		for _, a := range arr {
			t.Log(a)
		}
		t.Fatal(s)
	}
}
