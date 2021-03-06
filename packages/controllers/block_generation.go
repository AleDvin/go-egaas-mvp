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
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
)

type forgingPage struct {
	Lang         map[string]string
	Title        string
	CountSign    int
	CountSignArr []int
	SignData     string
	ShowSignData bool
	MyWalletData map[string]string
	WalletId     int64
	CitizenId    int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) Forging() (string, error) {

	txType := "DLTChangeHostVote"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	MyWalletData, err := c.OneRow("SELECT host, address_vote FROM dlt_wallets WHERE wallet_id = ?", c.SessWalletId).String()
	MyWalletData[`address`] = lib.AddressToString(uint64(c.SessWalletId))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("MyWalletData %v", MyWalletData)

	TemplateStr, err := makeTemplate("forging", "forging", &forgingPage{
		Lang:         c.Lang,
		MyWalletData: MyWalletData,
		Title:        "modalAnonym",
		ShowSignData: c.ShowSignData,
		SignData:     "",
		WalletId:     c.SessWalletId,
		CitizenId:    c.SessCitizenId,
		CountSignArr: c.CountSignArr,
		CountSign:    c.CountSign,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
