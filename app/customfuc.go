package app

import (
	"github.com/go-redis/redis"
	log2 "rulecat/utils/log"
	"sync"
)

/*

引入的配置，在init函数里面初始化即可
规则添加方式如下：

  - field: conn.dip
    type: customf
    rule: CheckIP

然后，在handleMap里面注册一下就可以，
RegisterHandler("CheckIP", CheckIP)



*/

var (
	RedisClientIP  *redis.Client
	RedisClientDNS *redis.Client
	HandleMap      = make(map[string]func(interface{}) (bool, map[string]string))
	mu             sync.RWMutex
)

func init() {

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

	RegisterHandler("CheckIP", CheckIP)
	RegisterHandler("CheckDNS", CheckDNS)
	RegisterHandler("CheckProto", CheckProto)
}

func RegisterHandler(name string, handler func(interface{}) (bool, map[string]string)) {
	mu.Lock()
	defer mu.Unlock()
	HandleMap[name] = handler
}

func CheckIP(ip interface{}) (bool, map[string]string) {
	ipStr, ok := ip.(string)
	if !ok {
		return false, nil
	}

	val2, err := RedisClientIP.Get(ipStr).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		log2.Error.Println("Error fetching IP from Redis:", err)
		return false, nil
	}

	if val2 != "" {
		return true, map[string]string{
			"ip_tag": val2,
		}
	}
	return false, nil
}

func CheckDNS(dns interface{}) (bool, map[string]string) {
	dnsStr, ok := dns.(string)
	if !ok {
		return false, nil
	}

	val2, err := RedisClientDNS.Get(dnsStr).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		log2.Error.Println("Error fetching DNS from Redis:", err)
		return false, nil
	}

	if val2 != "" {
		return true, map[string]string{
			"dns_tag": val2,
		}
	}
	return false, nil
}

func CheckProto(proto interface{}) (bool, map[string]string) {

	protocolStr, ok := proto.(string)
	if !ok {
		return false, nil
	}
	if protocolStr == "UDP" {
		return true, map[string]string{
			"pro_tag": protocolStr,
		}
	}
	return false, nil

}
