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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) ParseInit() error {
	/*if p.States == nil {
		p.States = make(map[int64]string)
		all, err := p.GetAll(`SELECT id, name FROM system_states`, -1)
		if err != nil {
			return err
		}
		for _, v := range all {
			p.States[utils.StrToInt64(v["id"])] = v["name"]
		}
	}*/
	return nil
}

func (p *Parser) ParseBlock() error {
	/*
		Заголовок
		TYPE (0-блок, 1-тр-я)     1
		BLOCK_ID   				       4
		TIME       					       4
		WALLET_ID                         1-8
		state_id                         1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	p.BlockData = utils.ParseBlockHeader(&p.BinaryData)
	log.Debug("%v", p.BlockData)

	p.CurrentBlockId = p.BlockData.BlockId

	// Until then let it be. Get tables p_keys. then it is necessary to update only when you change tables
	allTables, err := p.GetAllTables()
	if err != nil {
		return utils.ErrInfo(err)
	}
	p.AllPkeys = make(map[string]string)
	for _, table := range allTables {
		log.Debug("%s", table)
		col, err := p.GetFirstColumnName(table)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("%s", col)
		p.AllPkeys[table] = col
	}
	if err := p.ParseInit(); err != nil {
		return utils.ErrInfo(err)
	}

	return nil
}
