package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)


type blockGenerationPage struct {
	Lang                  map[string]string
	Title                 string
	CountSign             int
	CountSignArr          []int
	SignData              string
	ShowSignData          bool
	MyWalletData		  map[string]string
	WalletId int64
	CitizenId int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) BlockGeneration() (string, error) {

	txType := "DLTChangeHostVote"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	MyWalletData, err := c.OneRow("SELECT hex(address) as address, host, hex(vote) as vote  FROM dlt_wallets WHERE wallet_id = ?", c.SessWalletId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("MyWalletData %v", MyWalletData);


	TemplateStr, err := makeTemplate("block_generation", "blockGeneration", &blockGenerationPage{
		Lang:                  c.Lang,
		MyWalletData:          MyWalletData,
		Title:                 "modalAnonym",
		ShowSignData:          c.ShowSignData,
		SignData:              "",
		WalletId: c.SessWalletId,
		CitizenId: c.SessCitizenId,
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}