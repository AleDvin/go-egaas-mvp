// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type newTablePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	CitizenId    int64
	StateId      int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	Global       string
}

func (c *Controller) NewTable() (string, error) {

	var err error

	global := c.r.FormValue("global")
	if global == "" {
		global = "0"
	}

	txType := "NewTable"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("new_table", "newTable", &newTablePage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		CitizenId:    c.SessCitizenId,
		StateId:      c.StateId,
		Global:       global,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
