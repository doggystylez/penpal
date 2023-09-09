package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetLatestBlock(url string, client *http.Client) (responseData Block, err error) {
	err = getByURLAndUnmarshal(&responseData, url+"/block", client)
	return responseData, err
}

func GetLatestHeight(url string, client *http.Client) (height string, err error) {
	block, err := GetLatestBlock(url, client)
	if err != nil {
		return "", err
	}
	return block.Result.Block.Header.Height, err
}

func GetLatestBlockTime(url string, client *http.Client) (chainID string, blockTime time.Time, err error) {
	block, err := GetLatestBlock(url, client)
	if err != nil {
		return "", time.Time{}, err
	}
	return block.Result.Block.Header.ChainID, block.Result.Block.Header.Time, err
}

func GetBlockFromHeight(height string, url string, client *http.Client) (responseData Block, err error) {
	err = getByURLAndUnmarshal(&responseData, url+"/block?height="+height, client)
	return responseData, err
}

func getByURLAndUnmarshal(data interface{}, url string, client *http.Client) (err error) {
	r := &strings.Reader{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, r)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &data)
	return err
}
