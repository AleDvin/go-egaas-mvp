package parser

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

/**
 * Откат таблиц rb_time_, которые были изменены транзакциями
 */
func (p *Parser) ParseDataRollbackFront(txcandidateBlock bool) error {

	// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
	binForSize := p.BinaryData
	var sizesSlice []int64
	for {
		txSize := utils.DecodeLength(&binForSize)
		if txSize == 0 {
			break
		}
		sizesSlice = append(sizesSlice, txSize)
		// удалим тр-ию
		utils.BytesShift(&binForSize, txSize)
		if len(binForSize) == 0 {
			break
		}
	}
	sizesSlice = utils.SliceReverse(sizesSlice)
	for i := 0; i < len(sizesSlice); i++ {
		// обработка тр-ий может занять много времени, нужно отметиться
		p.UpdDaemonTime(p.GoroutineName)
		// отделим одну транзакцию
		transactionBinaryData := utils.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
		// узнаем кол-во байт, которое занимает размер
		size_ := len(utils.EncodeLength(sizesSlice[i]))
		// удалим размер
		utils.BytesShiftReverse(&p.BinaryData, size_)
		p.TxHash = string(utils.Md5(transactionBinaryData))

		// инфа о предыдущем блоке (т.е. последнем занесенном)
		err := p.GetInfoBlock()
		if err != nil {
			return p.ErrInfo(err)
		}
		if txcandidateBlock {
			utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE hex(hash) = " + string(p.TxHash))
			affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				utils.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
		}
		affected, err := p.ExecSqlGetAffect("DELETE FROM rb_transactions WHERE hex(hash) = ?", p.TxHash)
		log.Debug("DELETE FROM rb_transactions WHERE hex(hash) = %s / affected = %d", p.TxHash, affected)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.dataType = utils.BytesToInt(p.TxSlice[1])
		//userId := p.TxSlice[3]
		MethodName := consts.TxTypes[p.dataType]
		err_ := utils.CallMethod(p, MethodName+"Init")
		if _, ok := err_.(error); ok {
			return p.ErrInfo(err_.(error))
		}
		err_ = utils.CallMethod(p, MethodName+"RollbackFront")
		if _, ok := err_.(error); ok {
			return p.ErrInfo(err_.(error))
		}
	}

	return nil
}


/**
 * Откат БД по блокам
 */
func (p *Parser) ParseDataRollback() error {

	p.dataPre()
	if p.dataType != 0 { // парсим только блоки
		return utils.ErrInfo(fmt.Errorf("incorrect dataType"))
	}
	var err error

	err = p.ParseBlock()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if len(p.BinaryData) > 0 {
		// вначале нужно получить размеры всех тр-ий, чтобы пройтись по ним в обратном порядке
		binForSize := p.BinaryData
		var sizesSlice []int64
		for {
			txSize := utils.DecodeLength(&binForSize)
			if txSize == 0 {
				break
			}
			sizesSlice = append(sizesSlice, txSize)
			// удалим тр-ию
			utils.BytesShift(&binForSize, txSize)
			if len(binForSize) == 0 {
				break
			}
		}
		sizesSlice = utils.SliceReverse(sizesSlice)
		for i := 0; i < len(sizesSlice); i++ {
			// обработка тр-ий может занять много времени, нужно отметиться
			p.UpdDaemonTime(p.GoroutineName)
			// отделим одну транзакцию
			transactionBinaryData := utils.BytesShiftReverse(&p.BinaryData, sizesSlice[i])
			// узнаем кол-во байт, которое занимает размер
			size_ := len(utils.EncodeLength(sizesSlice[i]))
			// удалим размер
			utils.BytesShiftReverse(&p.BinaryData, size_)
			p.TxHash = string(utils.Md5(transactionBinaryData))

			utils.WriteSelectiveLog("UPDATE transactions SET used=0, verified = 0 WHERE hex(hash) = " + string(p.TxHash))
			affect, err := p.ExecSqlGetAffect("UPDATE transactions SET used=0, verified = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				utils.WriteSelectiveLog(err)
				return p.ErrInfo(err)
			}
			utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			affected, err := p.ExecSqlGetAffect("DELETE FROM rb_transactions WHERE hex(hash) = ?", p.TxHash)
			log.Debug("DELETE FROM rb_transactions WHERE hex(hash) = %s / affected = %d", p.TxHash, affected)
			if err != nil {
				return p.ErrInfo(err)
			}
			// даем юзеру понять, что его тр-ия не в блоке
			err = p.ExecSql("UPDATE transactions_status SET block_id = 0 WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			// пишем тр-ию в очередь на проверку, авось пригодится
			dataHex := utils.BinToHex(transactionBinaryData)
			err = p.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", p.TxHash)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", p.TxHash, dataHex)
			if err != nil {
				return p.ErrInfo(err)
			}

			p.TxSlice, err = p.ParseTransaction(&transactionBinaryData)
			if err != nil {
				return p.ErrInfo(err)
			}
			p.dataType = utils.BytesToInt(p.TxSlice[1])
			MethodName := consts.TxTypes[p.dataType]
			err_ := utils.CallMethod(p, MethodName+"Init")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
			err_ = utils.CallMethod(p, MethodName+"Rollback")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
			err_ = utils.CallMethod(p, MethodName+"RollbackFront")
			if _, ok := err_.(error); ok {
				return p.ErrInfo(err_.(error))
			}
		}
	}
	return nil
}