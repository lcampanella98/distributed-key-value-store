package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

func Get(key string, addr string) (types.GetResponse, error) {
	params := url.Values{}
	params.Add("key", key)
	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s/get?%s", addr, queryString)
	resp, err := http.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Get: %v\n", err)
		return types.GetResponse{}, err
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error")
		return types.GetResponse{}, errors.New("internal server error on get")
	}
	defer resp.Body.Close()
	var res *types.GetResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return *res, nil
}

func Put(key string, value string, addr string) (types.PutResponse, error) {
	params := url.Values{}
	params.Add("key", key)
	params.Add("value", value)
	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s/put?%s", addr, queryString)
	resp, err := http.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Put: %v\n", err)
		return types.PutResponse{}, err
	}
	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error")
		return types.PutResponse{}, errors.New("internal server error on put")
	}
	defer resp.Body.Close()
	var res *types.PutResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return *res, nil
}
