package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

var client *http.Client

func Init(isInternal bool) {
	if isInternal {
		tr := &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   500 * time.Millisecond,
		}
	} else {
		tr := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   2 * time.Second,
		}
	}
}

func Get(key string, addr string) (types.GetResponse, error) {
	params := url.Values{}
	params.Add("key", key)
	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s/get?%s", addr, queryString)
	resp, err := client.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Get: %v\n", err)
		return types.GetResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error")
		return types.GetResponse{}, errors.New("internal server error on get")
	}
	var res *types.GetResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return *res, nil
}

func Put(key string, value string, addr string) (types.PutResponse, error) {
	return PutWithCoordinator(key, value, addr, "")
}

func PutWithCoordinator(key string, value string, addr string, coordinator string) (types.PutResponse, error) {
	params := url.Values{}
	params.Add("key", key)
	params.Add("value", value)
	params.Add("coordinator", coordinator)
	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s/put?%s", addr, queryString)
	resp, err := client.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Put: %v\n", err)
		return types.PutResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error")
		return types.PutResponse{}, errors.New("internal server error on put")
	}
	var res *types.PutResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return *res, nil
}

func Clear(addr string) error {
	fullURL := fmt.Sprintf("%s/clear", addr)
	resp, err := client.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Clear: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println("Internal server error")
		return errors.New("internal server error on clear")
	}
	return nil
}
