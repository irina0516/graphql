/*
Package rpc implements bridge to Lachesis full node API interface.

We recommend using local IPC for fast and the most efficient inter-process communication between the API server
and an Opera/Lachesis node. Any remote RPC connection will work, but the performance may be significantly degraded
by extra networking overhead of remote RPC calls.

You should also consider security implications of opening Lachesis RPC interface for a remote access.
If you considering it as your deployment strategy, you should establish encrypted channel between the API server
and Lachesis RPC interface with connection limited to specified endpoints.

We strongly discourage opening Lachesis RPC interface for unrestricted Internet access.
*/
package rpc

//go:generate tools/abigen.sh --abi ./contracts/abi/defi-tokens-registry.abi --pkg contracts --type DefiFMintTokenRegistry --out ./contracts/fmint_tokens.go

import (
	"fmt"
	"galaxy-graphql/internal/repository/rpc/contracts"
	"galaxy-graphql/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// DefiTokens resolves list of DeFi tokens available for the DeFi functions.
func (chain *ChainBridge) DefiTokens() ([]types.DefiToken, error) {
	// connect the contract
	contract, err := chain.fMintCfg.tokenRegistryContract()
	if err != nil {
		return nil, err
	}

	return chain.defiTokensList(contract)
}

// DefiTokenList creates a list of addresses / identifiers of all the ERC20 tokens
// involved with the fMint protocol.
func (chain *ChainBridge) DefiTokenList() ([]common.Address, error) {
	// connect the contract
	contract, err := chain.fMintCfg.tokenRegistryContract()
	if err != nil {
		return nil, err
	}

	return chain.defiTokenAddressList(contract.TokensCount, contract.TokensList)
}

// DefiToken loads details of a single DeFi token by it's address.
func (chain *ChainBridge) DefiToken(token *common.Address) (*types.DefiToken, error) {
	// connect the contract
	contract, err := chain.fMintCfg.tokenRegistryContract()
	if err != nil {
		return nil, err
	}

	return chain.defiTokenDetail(contract, token)
}

// defiTokenAddressList load list of addresses of tokens using given
// count function and item access function to do the loading.
func (chain *ChainBridge) defiTokenAddressList(
	fCount func(*bind.CallOpts) (*big.Int, error),
	fItem func(*bind.CallOpts, *big.Int) (common.Address, error),
) ([]common.Address, error) {
	// get the number of tokens in the reference aggregator
	count, err := fCount(nil)
	if err != nil {
		chain.log.Errorf("can not get tokens range; %s", err.Error())
		return nil, err
	}

	// make a container for tokens
	list := make([]common.Address, count.Uint64())
	index := new(big.Int)

	// load all the tokens in the contract
	for i := uint64(0); i < count.Uint64(); i++ {
		// read the indexed token from contract
		list[i], err = fItem(nil, index.SetUint64(i))
		if err != nil {
			chain.log.Errorf("token %d address not found; %s", i, err.Error())
			return nil, err
		}
	}

	return list, nil
}

// defiTokenDetail loads details of a token specified by the token address.
func (chain *ChainBridge) defiTokenDetail(contract *contracts.DefiFMintTokenRegistry, token *common.Address) (*types.DefiToken, error) {
	// get the token details
	tk, err := contract.Tokens(nil, *token)
	if err != nil {
		chain.log.Errorf("token %s not found; %s", token.String(), err.Error())
		return nil, err
	}

	// decode token details
	dt, err := decodeToken(token, tk)
	if err != nil {
		chain.log.Errorf("can not decode token %s; %s", token.String(), err.Error())
		return nil, err
	}

	return &dt, nil
}

// defiTokensList loads list of DeFi tokens from the smart contract.
func (chain *ChainBridge) defiTokensList(contract *contracts.DefiFMintTokenRegistry) ([]types.DefiToken, error) {
	// get tge list of addresses
	al, err := chain.defiTokenAddressList(contract.TokensCount, contract.TokensList)
	if err != nil {
		chain.log.Errorf("tokens list not available; %s", err.Error())
		return nil, err
	}

	// make a container for tokens
	list := make([]types.DefiToken, 0)

	// load all the tokens in the contract
	for i, addr := range al {
		// decode the token
		tk, err := chain.defiTokenDetail(contract, &addr)
		if err != nil {
			chain.log.Errorf("invalid token #%d; %s", i, err.Error())
			return nil, err
		}

		// add the token if it's still active
		if tk.IsActive {
			list = append(list, *tk)
		}
	}

	return list, nil
}

// decodeToken decodes the contract internal token representation
// into the API structure.
func decodeToken(addr *common.Address, tk struct {
	Id            *big.Int
	Name          string
	Symbol        string
	Decimals      uint8
	Logo          string
	Oracle        common.Address
	PriceDecimals uint8
	IsActive      bool
	CanDeposit    bool
	CanMint       bool
}) (types.DefiToken, error) {
	// do we have a valid token? fail if not
	if tk.Id == nil || 0 == tk.Id.Uint64() {
		return types.DefiToken{}, fmt.Errorf("token undefined")
	}

	// decode and return
	return types.DefiToken{
		Address:       *addr,
		Index:         hexutil.Uint64(tk.Id.Uint64()),
		Name:          tk.Name,
		Symbol:        tk.Symbol,
		LogoUrl:       tk.Logo,
		Decimals:      int32(tk.Decimals),
		PriceDecimals: int32(tk.PriceDecimals),
		IsActive:      tk.IsActive,
		CanDeposit:    tk.CanDeposit,
		CanMint:       tk.CanMint,
		CanBorrow:     false,
		CanTrade:      false,
	}, nil
}
