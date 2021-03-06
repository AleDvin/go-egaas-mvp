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

package smart

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	//"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	//"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type Contract struct {
	Name   string
	Called uint32
	Extend *map[string]interface{}
	Block  *script.Block
}

const (
	CALL_INIT  = 0x01
	CALL_FRONT = 0x02
	CALL_MAIN  = 0x04
)

var (
	smartVM *script.VM
)

/*
contract TXCitizenRequest {
	tx {
		PublicKey  bytes
		StateId    int
		FirstName  string
		MiddleName string "optional"
		LastName   string
	}
	func init {
//		Println("TXCitizenRequest init" + $FirstName, $citizen, "/", $wallet,"=", Balance($wallet))
	}
	func front {
//		Println("TXCitizenRequest front" + $MiddleName, StateParam($StateId, "citizenship_price"))
		if 10000 {
			error "not enough money"
		}
	}
	func main {

//		Println("TXCitizenRequest main" + $LastName)
	}

}

contract TXNewCitizen {
			func front {
//				Println("NewCitizen Front", $citizen, $state, $PublicKey )
			}
			func main {
//				Println("NewCitizen Main", $type, $citizen, $block )
//				DBInsert(Sprintf( "%d_citizens", $state), "public_key,block_id", $PublicKey, $block)
			}
}
		 map[string]string{
	//		`*parser.Parser`: `parser`,
	})

				"DBInsert":   DBInsert,
				"Balance":    Balance,
				"StateParam": StateParam,*/

func init() {
	smartVM = script.NewVM()

	smartVM.Extend(&script.ExtendData{map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
		"TxJson":  TxJson,
		"Float":   Float,
		"Money":   script.ValueToDecimal,
	}, map[string]string{
		`*smart.Contract`: `contract`,
	}})
}

// Compiles contract source code
func Compile(src string) error {
	return smartVM.Compile([]rune(src))
}

func CompileBlock(src string) (*script.Block, error) {
	return smartVM.CompileBlock([]rune(src))
}

func CompileEval(src string) error {
	return smartVM.CompileEval(src)
}

func EvalIf(src string, extend *map[string]interface{}) (bool, error) {
	return smartVM.EvalIf(src, extend)
}

func FlushBlock(root *script.Block) {
	smartVM.FlushBlock(root)
}

func Extend(ext *script.ExtendData) {
	smartVM.Extend(ext)
}

func Run(block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	rt := smartVM.RunInit()
	return rt.Run(block, params, extend)
}

// Returns true if the contract exists
func GetContract(name string /*, data interface{}*/) *Contract {
	obj, ok := smartVM.Objects[name]
	//	fmt.Println(`Get`, ok, obj, obj.Type, script.OBJ_CONTRACT)
	if ok && obj.Type == script.OBJ_CONTRACT {
		return &Contract{Name: name, Block: obj.Value.(*script.Block)}
	}
	return nil
}

// Returns true if the contract exists
func GetContractById(id int32 /*, p *Parser*/) *Contract {
	idcont := id // - CNTOFF
	if len(smartVM.Children) <= int(idcont) || smartVM.Children[idcont].Type != script.OBJ_CONTRACT {
		return nil
	}
	return &Contract{Name: smartVM.Children[idcont].Info.(*script.ContractInfo).Name,
		/*parser: p,*/ Block: smartVM.Children[idcont]}
}

func (contract *Contract) GetFunc(name string) *script.Block {
	if block, ok := (*contract).Block.Objects[name]; ok && block.Type == script.OBJ_FUNC {
		return block.Value.(*script.Block)
	}
	return nil
}

func TxJson(contract *Contract) string {
	lines := make([]string, 0)
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		switch fitem.Type.String() {
		case `string`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `int64`:
			lines = append(lines, fmt.Sprintf(`"%s": %d`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `[]uint8`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name,
				hex.EncodeToString((*(*contract).Extend)[fitem.Name].([]byte))))
		}
	}
	return `{` + strings.Join(lines, ",\r\n") + `}`
}

func Float(v interface{}) (ret float64) {
	switch value := v.(type) {
	case int64:
		ret = float64(value)
	case string:
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			ret = val
		}
	}
	return
}

// Pre-defined functions
/*
func CheckAmount() {
	amount, err := p.Single(`SELECT value FROM `+utils.Int64ToStr().TxVars[`state_code`]+`_state_parameters WHERE name = ?`, "citizenship_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	amountAndCommission, err := p.checkSenderDLT(amount, consts.COMMISSION)
	if err != nil {
		return p.ErrInfo(err)
	}
	if amount > amountAndCommission {
		return p.ErrInfo("incorrect amount")
	}
	// вычитаем из wallets_buffer
	// amount_and_commission взято из check_sender_money()
	err = p.updateWalletsBuffer(amountAndCommission)
	if err != nil {
		return p.ErrInfo(err)
	}

}
*/
