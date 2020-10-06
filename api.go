/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/types"
)

func param2Item(p *types.Params) (ret *model.Item) {
	now := time.Now().Unix()

	return &model.Item{
		ID:       p.ID,
		Driver:   p.Driver,
		Endpoint: p.Endpoint,
		Content:  p.Payload,
		CreateAt: now,
		NextAt:   now,
		Tried:    0,
		State:    types.PENDING,
	}
}

// DefaultScheduler retries every minute at first 10 tries, and doubles wait time each time
func DefaultScheduler(driver, notifyID string, lastExec time.Time, tried uint32) (next time.Time, stop bool) {
	next = lastExec.Add(time.Minute)
	if tried > 10 {
		delta := int64(math.Pow(float64(tried-10), 2.0))
		if x := math.MinInt64 / int64(time.Minute); delta > x {
			delta = x
		}
		next = lastExec.Add(time.Duration(delta) * time.Minute)
	}

	return
}

// APIServer defines an API server to accept notification-sending requests.
//
// API format
//
// Parameters are passed in JSON format using HTTP POST request. "Content-Type"
// header is ignored. The result of request is returned in HTTP status code.
//
// API Endpoints
//
//   - /send:       Send notification and retry automatically if not delivered. See
//                  types. Params struct for details of parameters.
//   - /sendOnce:   Send notification, does not retry. See Params struct for details
//                  of parameters.
//   - /resend:     Force resend a notification, does not retry. The only accpeted
//                  parameter is {"id": string}.
//   - /result:     Retrieve latest sending result. The only accpeted parameter is
//                  {"id": string}.
//   - /status:     Retrieve status of a notification, see types.Status for detail.
//                  It accepts only one parameter {"id": string}.
//   - /detail:     Retrieve detail of a notification, see types.Detail for detail.
//                  It accepts only one parameter {"id": string}.
//   - /delete:     Deletes a notification, does not interrupt if worker is sending
//                  it. The only accpeted parameter is {"id": string}.
//   - /clear:      Deletes outdated, finished jobs (status IN(SUCCESS, FAILED)).
//                  The only accepted parameter is {"before": unix timestamp}.
//   - /forceClear: Deletes all outdated jobs
//                  The only accepted parameter is {"before": unix timestamp}.
//
// Jobs allocated by a worker will not be deleted by /delete, /clear nor /forceClear.
type APIServer interface {
	// register supported drivers, you *MUST* register all needed drivers
	// before starting server.
	Register(types.Driver)
	// start the api server and bind it to addr. It also starts internal worker
	// to send notification.
	Start() error
	// start the api server and bind it to addr, with basic TLS settings. It
	// also starts internal worker to send notification.
	StartTLS(certFile, keyFile string) error
	// gracefully shutdown the api server and internal worker.
	Shutdown(ctx context.Context) (err error)
	// returns the http.Server so you can customize it. Do not start it by
	// yourself, or internal worker will not start.
	GetHTTPServer() (ret *http.Server)
}

// NewAPI creates an APIServer
func NewAPI(opt SenderOptions) (ret APIServer, err error) {
	s, err := newSender(opt)
	if err != nil {
		return
	}
	x := &api{
		srv:    &http.Server{},
		sender: s,
		DBDrv:  opt.DBDrv,
	}
	x.srv.Handler = x.getMux()
	return x, nil
}

type api struct {
	srv    *http.Server
	sender sender
	model.DBDrv
}

func (a *api) Register(d types.Driver) {
	a.sender.Register(d)
}
