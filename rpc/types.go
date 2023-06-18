package rpc

import (
	"net/http"
)

type (
	Client struct {
		Client *http.Client
		Url    string
	}

	Block struct {
		Error  interface{} `json:"error"`
		Result struct {
			Block struct {
				Header struct {
					ChainID string `json:"chain_id"`
					Height  string `json:"height"`
				} `json:"header"`
				LastCommit struct {
					Signatures []struct {
						ValidatorAddress string `json:"validator_address"`
					} `json:"signatures"`
				} `json:"last_commit"`
			} `json:"block"`
		} `json:"result"`
	}
)
