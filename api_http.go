/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"context"
	"net/http"
	"sync"
)

func (a *api) GetHTTPServer() (ret *http.Server) {
	return a.srv
}

func (a *api) getMux() (ret *http.ServeMux) {
	ret = &http.ServeMux{}
	ret.HandleFunc("/send", a.sendH)
	ret.HandleFunc("/sendOnce", a.sendOnceH)
	ret.HandleFunc("/resend", a.resendH)
	ret.HandleFunc("/result", a.resultH)
	ret.HandleFunc("/status", a.statusH)
	ret.HandleFunc("/detail", a.detailH)
	ret.HandleFunc("/delete", a.deleteH)
	ret.HandleFunc("/clear", a.clearH)
	ret.HandleFunc("/forceClear", a.forceClearH)

	return
}

func (a *api) Start() (err error) {
	go a.sender.Start()
	return a.srv.ListenAndServe()
}

func (a *api) StartTLS(certFile, keyFile string) (err error) {
	go a.sender.Start()
	return a.srv.ListenAndServeTLS(certFile, keyFile)
}

func (a *api) Shutdown(ctx context.Context) (err error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		a.sender.Stop(ctx)
		wg.Done()
	}()

	err = a.srv.Shutdown(ctx)
	wg.Wait()
	return
}
