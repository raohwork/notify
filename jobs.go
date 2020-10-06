/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package notify

import "sync"

type jobCtrl struct {
	jobIDs []string
	size   uint16
	sync.RWMutex
}

func newJobCtl(threads uint16) (ret *jobCtrl) {
	return &jobCtrl{
		jobIDs: make([]string, threads),
		size:   threads,
	}
}

func (j *jobCtrl) set(tid uint16, jid string) {
	j.jobIDs[tid] = jid
}

func (j *jobCtrl) lset(tid uint16, jid string) {
	j.Lock()
	defer j.Unlock()

	j.jobIDs[tid] = jid
}

func (j *jobCtrl) list() (ret []string) {
	j.RLock()
	defer j.RUnlock()

	ret = append(ret, j.jobIDs...)
	return
}

func (j *jobCtrl) rawList() (ret []string) {
	ret = append(ret, j.jobIDs...)
	return
}
