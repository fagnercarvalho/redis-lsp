package client

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Users []string
	Keys []string

	client *redis.Client
}

func New(address string, username string, password string, db int, dbCache bool) (Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: address,
		Username: username,
		Password: password,
		DB: db,
	})

	c := Redis{client:client}

	// TODO: run MONITOR command to stream SET/DEL/ACL SETUSER/ACL DELUSER commands to update users/keys lists

	if dbCache {
		users, err := c.getUsers()
		if err != nil {
			return c, err
		}

		keys, err := c.getKeys()
		if err != nil {
			return c, err
		}

		c.Users = users
		c.Keys = keys
	} else {
		c.Users = []string {}
		c.Keys = []string {}
	}

	return c, nil
}

func (r Redis) ExecuteCommand(ctx context.Context, command []interface{}) (interface{}, error) {
	return r.client.Do(ctx, command...).Result()
}

func (r Redis) getUsers() ([]string, error) {
	val, err := r.client.Do(context.Background(), "ACL", "USERS").Result()
	if err != nil {
		return nil, err
	}

	slice := val.([]interface{})

	result := make([]string, len(slice))
	for _, v := range slice {
		result = append(result, v.(string))
	}

	return result, nil
}

func (r Redis) getKeys() ([]string, error) {
	val, err := r.client.Do(context.Background(), "SCAN", "0", "COUNT", "50").Result()
	if err != nil {
		return nil, err
	}

	// The scan will return a slice containing the index to the next page and a slice of results.
	// We only want the slice of results.
	// Example: [15 [4 777 2 test2 1 test 7 6 8 3 5]]
	slice := val.([]interface{})[1].([]interface{})

	result := make([]string, len(slice))
	for _, v := range slice {
		result = append(result, v.(string))
	}

	return result, nil
}