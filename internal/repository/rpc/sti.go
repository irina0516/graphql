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

//go:generate tools/abigen.sh --abi ./contracts/abi/st_info.abi --pkg contracts --type StakerInfoContract --out ./contracts/staker_info.go

import (
	"encoding/json"
	"fmt"
	"galaxy-graphql/internal/repository/rpc/contracts"
	"galaxy-graphql/internal/types"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// stiRequestTimeout is number of seconds we wait for the staker information request to finish.
const stiRequestTimeout = 2 * time.Second

// stiNameCheckRegex is the expression used to check for staker name validity
var stiNameCheckRegex = regexp.MustCompile(`^[\w\d\s.\-_'$()]+$`)

// StakerInfo extracts an extended staker information from smart contact by their id.
func (chain *ChainBridge) StakerInfo(id *hexutil.Big) (*types.StakerInfo, error) {
	if id == nil {
		return nil, fmt.Errorf("validator ID not given")
	}

	// keep track of the operation
	chain.log.Debugf("loading staker information for staker #%d", id.ToInt().Uint64())

	// instantiate the contract and display its name
	contract, err := contracts.NewStakerInfoContract(chain.sfcConfig.StiContract, chain.eth)
	if err != nil {
		chain.log.Criticalf("failed to instantiate STI contract: %v", err)
		return nil, err
	}

	// call for data
	stUrl, err := contract.GetInfo(nil, (*big.Int)(id))
	if err != nil {
		chain.log.Errorf("failed to get the staker information: %v", err)
		return nil, err
	}

	// var url string
	if len(stUrl) == 0 {
		chain.log.Debugf("no information for staker #%d", id.ToInt().Uint64())
		return nil, nil
	}

	// try to download JSON for the info
	return chain.downloadStakerInfo(stUrl)
}

// downloadStakerInfo tries to download staker information from the given URL address.
func (chain *ChainBridge) downloadStakerInfo(stUrl string) (*types.StakerInfo, error) {
	// log what we are about to do
	chain.log.Debugf("downloading staker info address [%s]", stUrl)

	// make a http client
	cl := http.Client{Timeout: stiRequestTimeout}

	// prep request
	req, err := http.NewRequest(http.MethodGet, stUrl, nil)
	if err != nil {
		chain.log.Errorf("can not request given staker info; %s", err.Error())
		return nil, err
	}

	// be honest, set agent
	req.Header.Set("User-Agent", "Galaxy GraphQL API Server")

	// process the request
	res, err := cl.Do(req)
	if err != nil {
		chain.log.Errorf("can not download staker info; %s", err.Error())
		return nil, err
	}

	// read the response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		chain.log.Errorf("can not read staker info response; %s", err.Error())
		return nil, err
	}

	// try to parse
	var info types.StakerInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		chain.log.Errorf("invalid response for staker info; %s", err.Error())
		return nil, err
	}

	// do we have anything?
	if !chain.isValidStakerInfo(&info) {
		chain.log.Errorf("invalid response for staker info [%s]", stUrl)
		return nil, err
	}

	chain.log.Debugf("found staker [%s]", *info.Name)
	return &info, nil
}

// isValidStakerInfo check if the staker information is valid and can be used.
func (chain *ChainBridge) isValidStakerInfo(info *types.StakerInfo) bool {
	// name must be available
	if nil == info.Name || 0 == len(*info.Name) || !stiNameCheckRegex.Match([]byte(*info.Name)) {
		chain.log.Error("staker name not valid")
		return false
	}

	// check the logo URL
	if !isValidStakerInfoUrl(info.LogoUrl, true) {
		chain.log.Error("staker logo URL not valid")
		return false
	}

	// check the website
	if !isValidStakerInfoUrl(info.Website, false) {
		chain.log.Error("staker website URL not valid")
		return false
	}

	// check the contact URL
	if !isValidStakerInfoUrl(info.Contact, false) {
		chain.log.Error("staker contact URL not valid")
		return false
	}
	return true
}

// isValidStakerInfoUrl validates the given URL address from the staker info.
func isValidStakerInfoUrl(addr *string, reqHttps bool) bool {
	// do we even have an URL; it's ok if not
	if nil == addr || 0 == len(*addr) {
		return true
	}

	// try to decode the address
	u, err := url.ParseRequestURI(*addr)
	if err != nil || u.Scheme == "" || (reqHttps && u.Scheme != "https") {
		return false
	}
	return true
}
