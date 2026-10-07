package eth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Dial(rpc string) (*ethclient.Client, error) {
	if rpc == "" {
		return nil, nil
	}
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, fmt.Errorf("dial ethereum rpc: %w", err)
	}
	return client, nil
}
