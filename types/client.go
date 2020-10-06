/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Client defines golang client to call notify.APIServer
type Client interface {
	// Creates a new Client using current settings and different context. The
	// context is used with http.NewRequestWithContext to create *http.Request.
	With(ctx context.Context) (ret Client)

	// maps api endpoints to function
	Send(id string, driver string, ep string, payload interface{}) (err error)
	SendOnce(id string, driver string, ep string, payload interface{}) (err error)
	Resend(id string) (err error)
	Result(id string) (ret []byte, err error)
	Status(id string) (ret Status, err error)
	Detail(id string) (ret Detail, err error)
	Delete(id string) (err error)
	Clear(before time.Time) (err error)
	ForceClear(before time.Time) (err error)
}

// NewClient creates a Client
//
// server is address of the server in "http(s)://example.com:1234" format (no path)
func NewClient(server string, hc *http.Client) (ret Client) {
	if hc == nil {
		hc = &http.Client{}
	}

	return &client{host: server, hc: hc, ctx: context.Background()}
}

type client struct {
	host string
	hc   *http.Client
	ctx  context.Context
}

func (c *client) With(ctx context.Context) (ret Client) {
	return &client{
		host: c.host,
		hc:   c.hc,
		ctx:  ctx,
	}
}

func (c *client) exec(path string, data interface{}) (err error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(
		c.ctx, "POST", c.host+path, bytes.NewReader(buf),
	)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("failed to call %s: %d", path, resp.StatusCode)
	}
	return
}

func (c *client) query(path string, data, ret interface{}) (err error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(
		c.ctx, "POST", c.host+path, bytes.NewReader(buf),
	)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("failed to call %s: %d", path, resp.StatusCode)
		return
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(ret)
	return
}

func (c *client) Send(id string, driver string, ep string, payload interface{}) (err error) {
	data := map[string]interface{}{
		"id":       id,
		"type":     driver,
		"endpoint": ep,
		"payload":  payload,
	}

	return c.exec("/send", data)
}
func (c *client) SendOnce(id string, driver string, ep string, payload interface{}) (err error) {
	data := map[string]interface{}{
		"id":       id,
		"type":     driver,
		"endpoint": ep,
		"payload":  payload,
	}

	return c.exec("/sendOnce", data)
}
func (c *client) Resend(id string) (err error) {
	data := map[string]interface{}{"id": id}
	return c.exec("/resend", data)
}
func (c *client) Result(id string) (ret []byte, err error) {
	data := map[string]interface{}{"id": id}
	buf, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(
		c.ctx, "POST", c.host+"/result", bytes.NewReader(buf),
	)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	ret, err = ioutil.ReadAll(resp.Body)
	return
}
func (c *client) Status(id string) (ret Status, err error) {
	data := map[string]interface{}{"id": id}
	err = c.query("/status", data, &ret)
	return
}
func (c *client) Detail(id string) (ret Detail, err error) {
	data := map[string]interface{}{"id": id}
	err = c.query("/detail", data, &ret)
	return
}
func (c *client) Delete(id string) (err error) {
	data := map[string]interface{}{"id": id}
	return c.exec("/delete", data)
}
func (c *client) Clear(before time.Time) (err error) {
	data := map[string]interface{}{"before": before.Unix()}
	return c.exec("/clear", data)
}
func (c *client) ForceClear(before time.Time) (err error) {
	data := map[string]interface{}{"before": before.Unix()}
	return c.exec("/forceClear", data)
}
