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

package parser

import (
	"errors"
	"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) TxParser(hash, binaryTx []byte, myTx bool) error {

	// проверим, нет ли несовместимых тр-ий
	// 	&waitError  - значит просто откладываем обработку тр-ии на после того, как сформируются блок
	// $fatal_error - удаляем тр-ию, т.к. она некорректная

	var err error
	fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar := p.ClearIncompatibleTx(binaryTx, myTx)
	if len(fatalError) == 0 && len(waitError) == 0 {
		p.BinaryData = binaryTx
		err = p.ParseDataGate(false)
	}

	hashHex := utils.BinToHex(hash)
	if err != nil || len(fatalError) > 0 {
		p.DeleteQueueTx(hashHex) // удалим тр-ию из очереди
	}
	if err == nil && len(fatalError) > 0 {
		err = errors.New(fatalError)
	}
	if err == nil && len(waitError) > 0 {
		err = errors.New(waitError)
	}
	if err != nil {
		log.Error("err: %v", err)
		errText := fmt.Sprintf("%s", err)
		if len(errText) > 255 {
			errText = errText[:255]
		}
		fromGate, err := p.Single("SELECT from_gate FROM queue_tx WHERE hex(hash) = ?", hashHex).Int64()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("fromGate %d", fromGate)
		if fromGate == 0 {
			log.Debug("UPDATE transactions_status SET error = %s WHERE hex(hash) = %s", errText, hashHex)
			err = p.ExecSql("UPDATE transactions_status SET error = ? WHERE hex(hash) = ?", errText, hashHex)
			if err != nil {
				return utils.ErrInfo(err)
			}
		}
	} else {

		log.Debug("SELECT counter FROM transactions WHERE hex(hash) = ?", string(hashHex))
		utils.WriteSelectiveLog("SELECT counter FROM transactions WHERE hex(hash) = " + string(hashHex))
		counter, err := p.Single("SELECT counter FROM transactions WHERE hex(hash) = ?", hashHex).Int64()
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("counter: " + utils.Int64ToStr(counter))
		counter++
		utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hashHex))
		affect, err := p.ExecSqlGetAffect(`DELETE FROM transactions WHERE hex(hash) = ?`, hashHex)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))

		log.Debug("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES (%s, %s, %v, %v, %v, %v, %v, %v)", hashHex, utils.BinToHex(binaryTx), forSelfUse, txType, walletId, citizenId, thirdVar, counter)
		utils.WriteSelectiveLog("INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?)")
		// вставляем с verified=1
		err = p.ExecSql(`INSERT INTO transactions (hash, data, for_self_use, type, wallet_id, citizen_id, third_var, counter, verified) VALUES ([hex], [hex], ?, ?, ?, ?, ?, ?, 1)`, hashHex, utils.BinToHex(binaryTx), forSelfUse, txType, walletId, citizenId, thirdVar, counter)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("result insert")
		log.Debug("INSERT INTO transactions - OK")
		// удалим тр-ию из очереди (с verified=0)
		err = p.DeleteQueueTx(hashHex)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) DeleteQueueTx(hashHex []byte) error {

	log.Debug("DELETE FROM queue_tx WHERE hex(hash) = %s", hashHex)
	err := p.ExecSql("DELETE FROM queue_tx WHERE hex(hash) = ?", hashHex)
	if err != nil {
		return utils.ErrInfo(err)
	}
	// т.к. мы обрабатываем в queue_parser_tx тр-ии с verified=0, то после их обработки их нужно удалять.
	utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(hashHex) + " AND verified=0 AND used = 0")
	affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE hex(hash) = ? AND verified=0 AND used = 0", hashHex)
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
	return nil
}

func (p *Parser) AllTxParser() error {

	// берем тр-ии
	all, err := p.GetAll(`
			SELECT *
			FROM (
	              SELECT data,
	                         hash
	              FROM queue_tx
				UNION
				SELECT data,
							 hash
				FROM transactions
				WHERE verified = 0 AND
							 used = 0
			)  AS x
			`, -1)
	for _, data := range all {

		log.Debug("hash: %x", data["hash"])

		err = p.TxParser([]byte(data["hash"]), []byte(data["data"]), false)
		if err != nil {
			err0 := p.ExecSql(`INSERT INTO incorrect_tx (time, hash, err) VALUES (?, [hex], ?)`, utils.Time(), utils.BinToHex(data["hash"]), fmt.Sprintf("%s", err))
			if err0 != nil {
				log.Error("%v", utils.ErrInfo(err0))
			}
			return utils.ErrInfo(err)
		}
	}
	return nil
}
