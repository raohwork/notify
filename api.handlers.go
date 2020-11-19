/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

func (a *api) toItem(r *http.Request) (ret *model.Item, err error) {
	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	var p types.Params
	if err = dec.Decode(&p); err != nil {
		return
	}

	if p.ID == "" || p.Driver == "" {
		err = errors.New("missing required parameter")
		return
	}

	drv, ok := a.sender.driver(p.Driver)
	if !ok {
		err = errors.New("unsupported driver: " + p.Driver)
		return
	}

	if err = drv.Verify(p.Payload); err != nil {
		err = errors.New("unsupported payload")
		return
	}

	ret = param2Item(&p)
	return
}

func (a *api) sendH(w http.ResponseWriter, r *http.Request) {
	i, err := a.toItem(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	err = a.Create(i)
	if err != nil {
		// cannot save to db, might be duplicated or just db error
		w.WriteHeader(500)
	}
}

func (a *api) sendOnceH(w http.ResponseWriter, r *http.Request) {
	i, err := a.toItem(r)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	i.Tried = a.sender.maxRetry() - 1

	err = a.Create(i)
	if err != nil {
		// cannot save to db, might be duplicated or just db error
		w.WriteHeader(500)
		return
	}
}

func (a *api) resendH(w http.ResponseWriter, r *http.Request) {
	var p struct {
		ID string `json:"id"`
	}

	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	if p.ID == "" {
		// missing basic parameter
		w.WriteHeader(400)
		return
	}

	if err := a.Resend(p.ID, a.sender.maxRetry()); err != nil {
		// not found or just db error
		if _, ok := err.(*model.E404); ok {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}
}

func (a *api) statusH(w http.ResponseWriter, r *http.Request) {
	var p struct {
		ID string `json:"id"`
	}

	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	if p.ID == "" {
		// missing basic parameter
		w.WriteHeader(400)
		return
	}

	ret, err := a.Status(p.ID)
	if err != nil {
		if _, ok := err.(*model.E404); ok {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	buf, _ := json.Marshal(ret)
	w.Write(buf)
}

func (a *api) detailH(w http.ResponseWriter, r *http.Request) {
	var p struct {
		ID string `json:"id"`
	}

	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	if p.ID == "" {
		// missing basic parameter
		w.WriteHeader(400)
		return
	}

	ret, err := a.Detail(p.ID)
	if err != nil {
		if _, ok := err.(*model.E404); ok {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	buf, _ := json.Marshal(ret)
	w.Write(buf)
}

func (a *api) resultH(w http.ResponseWriter, r *http.Request) {
	var p struct {
		ID string `json:"id"`
	}

	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	if p.ID == "" {
		// missing basic parameter
		w.WriteHeader(400)
		return
	}

	resp, err := a.Result(p.ID)
	if err != nil {
		if _, ok := err.(*model.E404); ok {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(resp)
}

func (a *api) deleteH(w http.ResponseWriter, r *http.Request) {
	var p struct {
		ID string `json:"id"`
	}

	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	if p.ID == "" {
		// missing basic parameter
		w.WriteHeader(400)
		return
	}

	// TODO: log error
	if err := a.Delete(p.ID, a.sender.curID()); err != nil {
		if _, ok := err.(*model.E404); ok {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}
}

func (a *api) clearH(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)

	var p struct {
		Before int64 `json:"before"`
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	t := time.Unix(p.Before, 0)
	if err := a.Clear(t, a.sender.curID()); err != nil {
		w.WriteHeader(500)
	}
}

func (a *api) forceClearH(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer io.Copy(ioutil.Discard, r.Body)

	var p struct {
		Before int64 `json:"before"`
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		// incorrect format
		w.WriteHeader(400)
		return
	}

	t := time.Unix(p.Before, 0)
	if err := a.ForceClear(t, a.sender.curID()); err != nil {
		w.WriteHeader(500)
	}
}
