package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

var client *http.Client
var highTimeoutClient *http.Client

func initHighTimeoutInternalClient() {
	tr := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
	}
	highTimeoutClient = &http.Client{
		Transport: tr,
		Timeout:   2 * time.Second,
	}
}

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
	initHighTimeoutInternalClient()
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
	var res types.GetResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
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

	if resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("Unsuccessful put: %s\n", resp.Status)
		bodyBytes, err := io.ReadAll(resp.Body)
		var errorText string = resp.Status + " Reason: "
		if err != nil {
			fmt.Println("Could not read error body:", err)
			errorText += "(Could not read error body)"
		} else {
			errorText += string(bodyBytes)
		}

		fmt.Println(errorText)

		return types.PutResponse{}, errors.New(errorText)
	}
	var res types.PutResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
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

func Health(addr string) error {
	fullURL := fmt.Sprintf("%s/health", addr)
	resp, err := client.Get(fullURL)

	if err != nil {
		// fmt.Printf("Health check error for node at %s: %v\n", addr, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// fmt.Printf("Health check failed for node at %s: %v\n", addr, resp.Status)
		return errors.New("Health check failed: " + resp.Status)
	}
	return nil

}

func Kill(addr string) error {
	fullURL := fmt.Sprintf("%s/kill", addr)
	resp, err := client.Get(fullURL)

	if err != nil {
		fmt.Printf("Kill node error for node at %s: %v\n", addr, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Kill node failed for node at %s: %v\n", addr, resp.Status)
		return errors.New("Kill node failed: " + resp.Status)
	}
	return nil

}

func Repair(node string, addr string) (types.RepairResponse, error) {
	params := url.Values{}
	params.Add("node", node)
	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s/repair?%s", addr, queryString)
	// use high-timeout client
	resp, err := highTimeoutClient.Get(fullURL)

	if err != nil {
		fmt.Printf("Error in client Repair: %v\n", err)
		return types.RepairResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("Unsuccessful Repair: %s\n", resp.Status)
		bodyBytes, err := io.ReadAll(resp.Body)
		var errorText string = resp.Status + " Reason: "
		if err != nil {
			fmt.Println("Could not read error body:", err)
			errorText += "(Could not read error body)"
		} else {
			errorText += string(bodyBytes)
		}

		fmt.Println(errorText)

		return types.RepairResponse{}, errors.New(errorText)
	}
	var res types.RepairResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
}
