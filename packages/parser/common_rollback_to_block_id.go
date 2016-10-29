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
	"database/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) RollbackToBlockId(blockId int64) error {

	/*err := p.ExecSql("SET GLOBAL net_read_timeout = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("SET GLOBAL max_connections  = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}*/
	/*err := p.RollbackTransactions()
	if err != nil {
		return p.ErrInfo(err)
	}*/
	err := p.ExecSql("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	if err != nil {
		utils.WriteSelectiveLog(err)
		return p.ErrInfo(err)
	}

	// откатываем наши блоки
	var blocks []map[string][]byte
	rows, err := p.Query(p.FormatQuery("SELECT id, data FROM block_chain WHERE id > ? ORDER BY id DESC"), blockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	parser := new(Parser)
	parser.DCDB = p.DCDB
	for rows.Next() {
		var data, id []byte
		err = rows.Scan(&id, &data)
		if err != nil {
			rows.Close()
			return p.ErrInfo(err)
		}
		blocks = append(blocks, map[string][]byte{"id": id, "data": data})
	}
	rows.Close()
	for _, block := range blocks {
		// Откатываем наши блоки до блока blockId
		parser.BinaryData = block["data"]
		err = parser.ParseDataRollback()
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.ExecSql("DELETE FROM block_chain WHERE id = ?", block["id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	var hash, data []byte
	err = p.QueryRow(p.FormatQuery("SELECT hash, data FROM block_chain WHERE id  =  ?"), blockId).Scan(&hash, &data)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	utils.BytesShift(&data, 1)
	block_id := utils.BinToDecBytesShift(&data, 4)
	time := utils.BinToDecBytesShift(&data, 4)
	size := utils.DecodeLength(&data)
	walletId := utils.BinToDecBytesShift(&data, size)
	CBID := utils.BinToDecBytesShift(&data, 1)
	err = p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, wallet_id = ?, state_id = ?", utils.BinToHex(hash), block_id, time, walletId, CBID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE config SET my_block_id = ?", block_id)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}