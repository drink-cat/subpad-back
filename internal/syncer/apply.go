package syncer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/model"
)

func applyLog(ctx context.Context, store *model.Store, chainID int, lg types.Log) error {
	if len(lg.Topics) == 0 {
		return nil
	}
	switch lg.Topics[0] {
	case launchABI.Events["TokenCreated"].ID:
		return applyTokenCreated(ctx, store, chainID, lg)
	case launchABI.Events["FeeCharged"].ID:
		return applyFeeCharged(ctx, store, chainID, lg)
	case launchABI.Events["SwapOnce"].ID:
		return applySwapOnce(ctx, store, chainID, lg)
	default:
		return nil
	}
}

func applyTokenCreated(ctx context.Context, store *model.Store, chainID int, lg types.Log) error {
	ev, err := decodeTokenCreated(lg)
	if err != nil {
		return err
	}
	row, err := store.TokenInfo.GetByPool(ctx, chainID, ev.PoolID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row, err = store.TokenInfo.FindPending(ctx, chainID, ev.Symbol)
	}
	if err != nil {
		return err
	}
	if row == nil {
		row = &model.TokenInfo{ChainID: chainID}
		fillToken(row, ev)
		if err = store.TokenInfo.Create(ctx, row); err != nil {
			return err
		}
	} else {
		fillToken(row, ev)
		if err = store.TokenInfo.Update(ctx, row); err != nil {
			return err
		}
	}
	slog.Info("applied event", "event", "TokenCreated", "chainId", chainID, "poolId", ev.PoolID, "token", ev.Token, "id", row.ID)
	return nil
}

func fillToken(row *model.TokenInfo, ev tokenCreated) {
	row.PoolID = ev.PoolID
	row.Creator = ev.Creator
	row.TokenAddr = ev.Token
	row.TokenName = ev.Name
	row.TokenSymbol = ev.Symbol
	row.QuoteTokenAddr = ev.QuoteToken
	row.QuoteTokenSymbol = ev.QuoteSymbol
	row.LaunchSupply = ev.LaunchSupply
	row.TickSpacing = ev.TickSpacing
}

func applyFeeCharged(ctx context.Context, store *model.Store, chainID int, lg types.Log) error {
	ev, err := decodeFeeCharged(lg)
	if err != nil {
		return err
	}
	txHash := lg.TxHash.Hex()
	logIndex := int(lg.Index)
	row := &model.FeeInfo{
		ChainID:    chainID,
		PoolID:     ev.PoolID,
		TxHash:     txHash,
		LogIndex:   logIndex,
		FeeType:    ev.FeeType,
		FeeToken:   ev.FeeToken,
		FeeDecimal: ev.FeeDecimal,
		FeeAmount:  ev.FeeAmount,
		FeeTo:      ev.FeeTo,
	}
	old, err := store.FeeInfo.GetByLog(ctx, chainID, txHash, logIndex)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err = store.FeeInfo.Create(ctx, row); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		row.ID = old.ID
		row.CreatedAt = old.CreatedAt
		if err = store.FeeInfo.Update(ctx, row); err != nil {
			return err
		}
	}
	slog.Info("applied event", "event", "FeeCharged", "chainId", chainID, "tx", txHash, "logIndex", logIndex, "feeType", ev.FeeType)
	return nil
}

func applySwapOnce(ctx context.Context, store *model.Store, chainID int, lg types.Log) error {
	ev, err := decodeSwapOnce(lg)
	if err != nil {
		return err
	}
	txHash := lg.TxHash.Hex()
	logIndex := int(lg.Index)
	row := &model.SwapInfo{
		ChainID:        chainID,
		PoolID:         ev.PoolID,
		TxHash:         txHash,
		LogIndex:       logIndex,
		Trader:         ev.Trader,
		IsBuy:          ev.IsBuy,
		TokenAddr:      ev.Token,
		TokenAmount:    ev.TokenAmount,
		TokenDecimal:   ev.TokenDecimal,
		QuoteTokenAddr: ev.QuoteToken,
		QuoteAmount:    ev.QuoteAmount,
		Fee:            ev.Fee,
		QuoteDecimal:   ev.QuoteDecimal,
		Price:          ev.Price,
	}
	old, err := store.SwapInfo.GetByLog(ctx, chainID, txHash, logIndex)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err = store.SwapInfo.Create(ctx, row); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		row.ID = old.ID
		row.CreatedAt = old.CreatedAt
		if err = store.SwapInfo.Update(ctx, row); err != nil {
			return err
		}
	}
	slog.Info("applied event", "event", "SwapOnce", "chainId", chainID, "tx", txHash, "logIndex", logIndex, "isBuy", ev.IsBuy)
	return nil
}
