/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package tgdrv provides a driver that send telegram message
package tgdrv

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/raohwork/notify/types"
)

// TGMarkdown is driver type of Markdown()
const TGMarkdown = "TGMarkdown"

func uri(token, ep string) (ret string) {
	return "https://api.telegram.org/bot" + token + "/" + ep
}

func validateToken(token string, cl *http.Client) (err error) {
	req, err := http.NewRequest("GET", uri(token, "getMe"), nil)
	if err != nil {
		return
	}
	resp, err := cl.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	type simpleUser struct {
		OK bool `json:"ok"`
	}
	var p simpleUser
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(&p); err != nil {
		return
	}

	if !p.OK {
		err = errors.New("invalid response")
		return
	}

	return
}

// Markdown creates a driver that send markdown formatted message in telegram
//
// The driver accepts string which is markdown formatted text.
//
// dest maps endpoint to chat/channel/user ids.
func Markdown(token string, dest map[string]int64, cl *http.Client) (ret types.Driver, err error) {
	if err = validateToken(token, cl); err != nil {
		err = errors.New("cannot validate telegram bot token: " + err.Error())
		return
	}

	ret = &md{
		token: token,
		dest:  dest,
		cl:    cl,
	}

	return
}

type md struct {
	token string
	dest  map[string]int64
	cl    *http.Client
}

func (t *md) Type() (ret string) {
	return TGMarkdown
}

func (t *md) Verify(data []byte) (err error) {
	_, err = t.extract(data)
	return
}

func (t *md) extract(data []byte) (msg string, err error) {
	if len(data) == 0 {
		err = errors.New("empty message")
	}

	err = json.Unmarshal(data, &msg)
	return
}

func (t *md) Send(ep string, content []byte) (resp []byte, err error) {
	cid, ok := t.dest[ep]
	if !ok {
		err = errors.New("unsupported dest: " + ep)
		return
	}

	txt, err := t.extract(content)
	if err != nil {
		return
	}

	val := url.Values{}
	val.Set("chat_id", strconv.FormatInt(cid, 10))
	val.Set("text", txt)
	val.Set("parse_mode", "MarkdownV2")

	req, err := http.NewRequest(
		"POST", uri(t.token, "sendMessage"),
		strings.NewReader(val.Encode()),
	)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := t.cl.Do(req)
	if err == nil || res.Body != nil {
		resp, _ = ioutil.ReadAll(res.Body)
		var x struct {
			OK   bool   `json:"ok"`
			Desc string `json:"description,omitempty"`
		}

		err = json.Unmarshal(resp, &x)
		if err == nil && !x.OK {
			err = errors.New("cannot send message: " + x.Desc)
		}
		res.Body.Close()
	}

	return
}
