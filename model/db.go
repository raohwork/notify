/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package model

import (
	"database/sql"
	"errors"
	"log"
)

// DrvBase is a helper to cache *sql.Stmt
type DrvBase struct {
	DB    *sql.DB
	stmts map[string]*sql.Stmt
}

// NewDrvBase creates a DrvBase
func NewDrvBase(db *sql.DB) (ret *DrvBase) {
	return &DrvBase{
		DB:    db,
		stmts: map[string]*sql.Stmt{},
	}
}

// Prepare caches a *sql.Stmt
//
// It supports error checking like this:
//
//     err = d.Prepare(query1, nil)
//     err = d.Prepare(query2, err)
//     err = d.Prepare(query3, err)
//     if err != nil {
//         return err
//     }
func (d *DrvBase) Prepare(qstr string, e error) (err error) {
	if e != nil {
		return e
	}
	if qstr == "" {
		return errors.New("empty sql statement!?")
	}

	if _, ok := d.stmts[qstr]; ok {
		return
	}

	stmt, err := d.DB.Prepare(qstr)
	if err != nil {
		log.Fatalf("cannot prepare %s: %s", qstr, err)
		return
	}

	d.stmts[qstr] = stmt
	return
}

// Stmt retrieves previously cached *sql.Stmt, returns nil if not found
//
// It is *RECOMMENDED* to store your sql query statement in constant or variable,
// and call Prepare()/Stmt() with that const/var. See package mysqldrv as a simple
// example, and package pgsqldrv for a even better example.
func (d *DrvBase) Stmt(qstr string) (ret *sql.Stmt) {
	return d.stmts[qstr]
}
