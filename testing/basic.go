package main

import (
	"fmt"

	"github.com/lcampanella98/distributed-key-value-store/internal/client"
	"github.com/lcampanella98/distributed-key-value-store/internal/types"
)

func TestBasic() {
	fmt.Println("Testing KV Store")
	addr := "http://localhost:8080"
	var getRes types.GetResponse
	var putRes types.PutResponse
	getRes, _ = client.Get("hi", addr)
	fmt.Println(getRes)
	putRes, _ = client.Put("hi", "John", addr)
	fmt.Println(putRes)
	getRes, _ = client.Get("hi", addr)
	fmt.Println(getRes)

	putRes, _ = client.Put("test", "Super", addr)
	fmt.Println(putRes)
	getRes, _ = client.Get("test", addr)
	fmt.Println(getRes)
}
