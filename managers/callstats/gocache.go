package callstats

func SetLocalCache(key string, value []byte) {
	err := callStatObj.Set(key, value, 10)
	if err != nil {
		return
	}
}

func GetLocalCache(key string) interface{} {
	value, err := callStatObj.Get(key)
	if err == nil {
		return value
	}
	return nil
}
