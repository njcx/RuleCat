package app

import (
	"github.com/go-redis/redis"
)

/*

引入的配置，在init函数里面初始化即可
规则添加方式如下：

  - field: conn.dip
    type: customf
    rule: CheckIP

然后，在handleMap里面注册一下就可以，
HandleMap["CheckIP"] = CheckIP
就像这样~


*/

var RedisClientIP *redis.Client
var RedisClientDNS *redis.Client

type handle func(name interface{}) (bool, map[string]string)

var HandleMap map[string]handle

func init() {
	HandleMap = make(map[string]handle)
	HandleMap["CheckIP"] = CheckIP
	HandleMap["CheckDNS"] = CheckDNS

	RedisClientIP = redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       0,
	})

	RedisClientDNS = redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       1,
	})
}

func CheckIP(ip interface{}) (bool, map[string]string) {

	val2, _ := RedisClientIP.Get(ip.(string)).Result()

	if len(val2) > 1 {
		return true, map[string]string{
			"ip_tag": val2}
	}
	return false, nil
}

func CheckDNS(dns interface{}) (bool, map[string]string) {

	val2, _ := RedisClientDNS.Get(dns.(string)).Result()

	if len(val2) > 1 {
		return true, map[string]string{
			"dns_tag": val2}
	}
	return false, nil
}
