package main

func (config *Config) CacheGet(key string) string {
	return ""
}

func (config *Config) CacheSet(key, value string) error {
	rows, err := config.Conn.Query(`SELECT value FROM json_cache
where json_cache.key = ` + key)
	defer rows.Close()
	if err != nil {
		return err
	}
	return nil
}
