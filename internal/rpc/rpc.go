package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func New() (client Client) {
	return Client{
		Client: &http.Client{
			Timeout: time.Second * 2,
		},
	}
}

func GetLastestHeight(client Client) (chainID string, height string, err error) {
	block, err := getLatestBlock(client)
	return block.Result.Block.Header.ChainID, block.Result.Block.Header.Height, err
}

func getLatestBlock(client Client) (responseData Block, err error) {
	err = getByUrlAndUnmarshall(client.Client, client.Url+"/block", &responseData)
	return
}

func GetBlockFromHeight(client Client, height string) (responseData Block, err error) {
	err = getByUrlAndUnmarshall(client.Client, client.Url+"/block?height="+height, &responseData)
	return
}

func getByUrlAndUnmarshall(client *http.Client, url string, data interface{}) (err error) {
	r := &strings.Reader{}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, r)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &data)
	return
}
