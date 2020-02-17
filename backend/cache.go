package main

import (
	"context"
	"encoding/json"
)

func (config *Config) CacheGet(key string) (ret map[string]interface{}, err error) {
	str, err := config.CacheGetStr(key)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(str), &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (config *Config) CacheSet(key string, val map[string]interface{}) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = config.CacheSetStr(key, string(b))
	return err
}

func (config *Config) CacheGetStr(key string) (string, error) {
	rows, err := config.conn.Query(
		context.Background(),
		"SELECT value FROM json_cache WHERE \"key\"='"+key+"'")
	defer rows.Close()
	if err != nil {
		return "{}", err
	}
	for rows.Next() {
		var str string
		rows.Scan(&str)
		return str, nil
	}
	return "{}", nil
}

func (config *Config) CacheSetStr(key, value string) error {
	_, err := config.conn.Exec(
		context.Background(),
		"INSERT INTO json_cache (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE set value = $2", key, value,
	)
	if err != nil {
		return err
	}
	return nil
}
