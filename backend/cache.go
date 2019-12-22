package main

func (config *Config) CacheGet(key string) (string, error) {
	rows, err := config.Conn.Query(`SELECT value from json_cache where json_cache.key = ` + key)
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

func (config *Config) CacheSet(key, value string) error {
	var a [][]interface{}
	a = append(a, []interface{}{
		key,
		value,
	})
	_, err := config.Conn.Exec(
		"INSERT INTO json_cache (key, value) VALUES ($1, $2)", key, value,
	)
	if err != nil {
		return err
	}
	return nil
}
