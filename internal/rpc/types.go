package rpc

import (
	"time"
)

type (
	Block struct {
		Error  interface{} `json:"error"`
		Result struct {
			Block struct {
				Header struct {
					ChainID string    `json:"chain_id"`
					Height  string    `json:"height"`
					Time    time.Time `json:"time"`
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
