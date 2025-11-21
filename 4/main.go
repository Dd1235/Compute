package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func main() {
	fmt.Println("Look into using Redis")
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// every IO operation should receive a context, so that you can cancel it, and propogate cancellations.

	// use a root context that won't be cancelled

	// it is top level, not tied to any request

	ping, err := client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(ping)

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	jsonString, err := json.Marshal(Person{
		Name: "Alice",
		Age:  20,
	})

	if err != nil {
		fmt.Printf("failed to marshal: %s", err.Error())
		return
	}

	// set the key foo to bar, with live forever, no expiration
	err = client.Set(context.Background(), "person", jsonString, 0).Err()
	if err != nil {
		fmt.Printf("Failed to set the value in the redis instance: %s", err.Error())
		return
	}

	val, err := client.Get(context.Background(), "person").Result()
	if err != nil {
		fmt.Printf("Failed to get the value from the redis instance for the key person")
	}
	fmt.Printf("Successfully fetched the value of foo as %v", val)

}

// docker pull redis
//  docker run --name redis-test-i
// nstance -p 6379:6379 -d redis

// docker ps
