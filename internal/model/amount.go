package model

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Amount 是链上 uint256 的十进制。JSON 仍按数字读写，数据库用字符串，避免超过 bigint。
type Amount string

func NewAmount(n int64) Amount {
	return Amount(strconv.FormatInt(n, 10))
}

func NewAmountBig(n *big.Int) (Amount, error) {
	if n == nil || n.Sign() < 0 {
		return "", fmt.Errorf("amount is invalid")
	}
	return Amount(n.String()), nil
}

func (a Amount) Equal(n int64) bool {
	if a == "" {
		return n == 0
	}
	return string(a) == strconv.FormatInt(n, 10)
}

func (a Amount) String() string {
	if a == "" {
		return "0"
	}
	return string(a)
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Amount) UnmarshalJSON(b []byte) error {
	if a == nil {
		return fmt.Errorf("amount is nil")
	}
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*a = "0"
		return nil
	}
	s = strings.Trim(s, `"`)
	if s == "" {
		*a = "0"
		return nil
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.Sign() < 0 {
		return fmt.Errorf("amount %s", s)
	}
	*a = Amount(n.String())
	return nil
}
