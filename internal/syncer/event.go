package syncer

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/drink-cat/subpad-back/internal/model"
)

const launchABIJSON = `[
  {"type":"event","name":"TokenCreated","inputs":[
    {"name":"poolId","type":"bytes32","indexed":true},
    {"name":"creator","type":"address","indexed":true},
    {"name":"token","type":"address","indexed":true},
    {"name":"tokenName","type":"string","indexed":false},
    {"name":"tokenSymbol","type":"string","indexed":false},
    {"name":"quoteToken","type":"address","indexed":false},
    {"name":"quoteTokenSymbol","type":"string","indexed":false},
    {"name":"launchSupply","type":"uint256","indexed":false},
    {"name":"tickSpacing","type":"int24","indexed":false}
  ]},
  {"type":"event","name":"FeeCharged","inputs":[
    {"name":"poolId","type":"bytes32","indexed":true},
    {"name":"feeType","type":"uint8","indexed":false},
    {"name":"feeToken","type":"address","indexed":true},
    {"name":"feeDecimal","type":"uint8","indexed":false},
    {"name":"feeAmount","type":"uint256","indexed":false},
    {"name":"feeTo","type":"address","indexed":true}
  ]},
  {"type":"event","name":"SwapOnce","inputs":[
    {"name":"poolId","type":"bytes32","indexed":true},
    {"name":"trader","type":"address","indexed":true},
    {"name":"isBuy","type":"bool","indexed":false},
    {"name":"token","type":"address","indexed":true},
    {"name":"tokenAmount","type":"uint256","indexed":false},
    {"name":"tokenDecimal","type":"uint8","indexed":false},
    {"name":"quoteToken","type":"address","indexed":false},
    {"name":"quoteAmount","type":"uint256","indexed":false},
    {"name":"fee","type":"uint256","indexed":false},
    {"name":"quoteDecimal","type":"uint8","indexed":false},
    {"name":"price","type":"uint256","indexed":false}
  ]}
]`

var launchABI = mustLaunchABI()

func mustLaunchABI() abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(launchABIJSON))
	if err != nil {
		panic(err)
	}
	return parsed
}

type tokenCreated struct {
	PoolID       string
	Creator      string
	Token        string
	Name         string
	Symbol       string
	QuoteToken   string
	QuoteSymbol  string
	LaunchSupply model.Amount
	TickSpacing  int
}

type feeCharged struct {
	PoolID     string
	FeeType    string
	FeeToken   string
	FeeDecimal int
	FeeAmount  model.Amount
	FeeTo      string
}

type swapOnce struct {
	PoolID       string
	Trader       string
	IsBuy        bool
	Token        string
	TokenAmount  model.Amount
	TokenDecimal int
	QuoteToken   string
	QuoteAmount  model.Amount
	Fee          model.Amount
	QuoteDecimal int
	Price        model.Amount
}

func decodeTokenCreated(lg types.Log) (tokenCreated, error) {
	ev := launchABI.Events["TokenCreated"]
	if lg.Topics[0] != ev.ID {
		return tokenCreated{}, fmt.Errorf("not TokenCreated")
	}
	if len(lg.Topics) != 4 {
		return tokenCreated{}, fmt.Errorf("TokenCreated topics %d", len(lg.Topics))
	}
	data, err := unpackData(ev, lg.Data)
	if err != nil {
		return tokenCreated{}, fmt.Errorf("TokenCreated data: %w", err)
	}
	name, err := mapString(data, "tokenName")
	if err != nil {
		return tokenCreated{}, err
	}
	symbol, err := mapString(data, "tokenSymbol")
	if err != nil {
		return tokenCreated{}, err
	}
	quote, err := mapAddr(data, "quoteToken")
	if err != nil {
		return tokenCreated{}, err
	}
	quoteSymbol, err := mapString(data, "quoteTokenSymbol")
	if err != nil {
		return tokenCreated{}, err
	}
	supply, err := mapAmount(data, "launchSupply")
	if err != nil {
		return tokenCreated{}, err
	}
	tick, err := mapInt(data, "tickSpacing")
	if err != nil {
		return tokenCreated{}, err
	}
	return tokenCreated{
		PoolID:       lg.Topics[1].Hex(),
		Creator:      common.BytesToAddress(lg.Topics[2].Bytes()).Hex(),
		Token:        common.BytesToAddress(lg.Topics[3].Bytes()).Hex(),
		Name:         name,
		Symbol:       symbol,
		QuoteToken:   quote.Hex(),
		QuoteSymbol:  quoteSymbol,
		LaunchSupply: supply,
		TickSpacing:  tick,
	}, nil
}

func decodeFeeCharged(lg types.Log) (feeCharged, error) {
	ev := launchABI.Events["FeeCharged"]
	if lg.Topics[0] != ev.ID {
		return feeCharged{}, fmt.Errorf("not FeeCharged")
	}
	if len(lg.Topics) != 4 {
		return feeCharged{}, fmt.Errorf("FeeCharged topics %d", len(lg.Topics))
	}
	data, err := unpackData(ev, lg.Data)
	if err != nil {
		return feeCharged{}, fmt.Errorf("FeeCharged data: %w", err)
	}
	feeType, err := mapUint8(data, "feeType")
	if err != nil {
		return feeCharged{}, err
	}
	name, err := feeTypeName(feeType)
	if err != nil {
		return feeCharged{}, err
	}
	feeDecimal, err := mapUint8(data, "feeDecimal")
	if err != nil {
		return feeCharged{}, err
	}
	amount, err := mapAmount(data, "feeAmount")
	if err != nil {
		return feeCharged{}, err
	}
	return feeCharged{
		PoolID:     lg.Topics[1].Hex(),
		FeeType:    name,
		FeeToken:   common.BytesToAddress(lg.Topics[2].Bytes()).Hex(),
		FeeDecimal: int(feeDecimal),
		FeeAmount:  amount,
		FeeTo:      common.BytesToAddress(lg.Topics[3].Bytes()).Hex(),
	}, nil
}

func decodeSwapOnce(lg types.Log) (swapOnce, error) {
	ev := launchABI.Events["SwapOnce"]
	if lg.Topics[0] != ev.ID {
		return swapOnce{}, fmt.Errorf("not SwapOnce")
	}
	if len(lg.Topics) != 4 {
		return swapOnce{}, fmt.Errorf("SwapOnce topics %d", len(lg.Topics))
	}
	data, err := unpackData(ev, lg.Data)
	if err != nil {
		return swapOnce{}, fmt.Errorf("SwapOnce data: %w", err)
	}
	isBuy, err := mapBool(data, "isBuy")
	if err != nil {
		return swapOnce{}, err
	}
	tokenAmount, err := mapAmount(data, "tokenAmount")
	if err != nil {
		return swapOnce{}, err
	}
	tokenDecimal, err := mapUint8(data, "tokenDecimal")
	if err != nil {
		return swapOnce{}, err
	}
	quote, err := mapAddr(data, "quoteToken")
	if err != nil {
		return swapOnce{}, err
	}
	quoteAmount, err := mapAmount(data, "quoteAmount")
	if err != nil {
		return swapOnce{}, err
	}
	fee, err := mapAmount(data, "fee")
	if err != nil {
		return swapOnce{}, err
	}
	quoteDecimal, err := mapUint8(data, "quoteDecimal")
	if err != nil {
		return swapOnce{}, err
	}
	price, err := mapAmount(data, "price")
	if err != nil {
		return swapOnce{}, err
	}
	return swapOnce{
		PoolID:       lg.Topics[1].Hex(),
		Trader:       common.BytesToAddress(lg.Topics[2].Bytes()).Hex(),
		IsBuy:        isBuy,
		Token:        common.BytesToAddress(lg.Topics[3].Bytes()).Hex(),
		TokenAmount:  tokenAmount,
		TokenDecimal: int(tokenDecimal),
		QuoteToken:   quote.Hex(),
		QuoteAmount:  quoteAmount,
		Fee:          fee,
		QuoteDecimal: int(quoteDecimal),
		Price:        price,
	}, nil
}

func feeTypeName(v uint8) (string, error) {
	switch v {
	case 0:
		return model.FeeTypePlatform, nil
	case 1:
		return model.FeeTypeTokenCreator, nil
	case 2:
		return model.FeeTypeSubpad, nil
	default:
		return "", fmt.Errorf("fee type %d", v)
	}
}

func unpackData(ev abi.Event, data []byte) (map[string]any, error) {
	out := make(map[string]any)
	if err := ev.Inputs.NonIndexed().UnpackIntoMap(out, data); err != nil {
		return nil, err
	}
	return out, nil
}

func mapString(m map[string]any, key string) (string, error) {
	v, ok := m[key].(string)
	if !ok {
		return "", fmt.Errorf("%s type %T", key, m[key])
	}
	return v, nil
}

func mapAddr(m map[string]any, key string) (common.Address, error) {
	v, ok := m[key].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("%s type %T", key, m[key])
	}
	return v, nil
}

func mapBool(m map[string]any, key string) (bool, error) {
	v, ok := m[key].(bool)
	if !ok {
		return false, fmt.Errorf("%s type %T", key, m[key])
	}
	return v, nil
}

func mapUint8(m map[string]any, key string) (uint8, error) {
	switch v := m[key].(type) {
	case uint8:
		return v, nil
	case *big.Int:
		if v == nil || !v.IsUint64() || v.Uint64() > 255 {
			return 0, fmt.Errorf("%s %v", key, v)
		}
		return uint8(v.Uint64()), nil
	default:
		return 0, fmt.Errorf("%s type %T", key, m[key])
	}
}

func mapAmount(m map[string]any, key string) (model.Amount, error) {
	v, ok := m[key].(*big.Int)
	if !ok {
		return "", fmt.Errorf("%s type %T", key, m[key])
	}
	amount, err := model.NewAmountBig(v)
	if err != nil {
		return "", fmt.Errorf("%s: %w", key, err)
	}
	return amount, nil
}

func mapInt(m map[string]any, key string) (int, error) {
	v, ok := m[key].(*big.Int)
	if !ok || v == nil || !v.IsInt64() {
		return 0, fmt.Errorf("%s type %T", key, m[key])
	}
	n := v.Int64()
	if int64(int(n)) != n {
		return 0, fmt.Errorf("%s %s does not fit int", key, v)
	}
	return int(n), nil
}
