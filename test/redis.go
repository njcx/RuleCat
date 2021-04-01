package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
)

var redisdb *redis.Client

// 初始化连接
func initClient() (err error) {
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// defer redisdb.Close()
	_, err = redisdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

// 字符串操作

func redisstring() {
	// 存/取  字符串key
	err := redisdb.Set("score", 100, 0).Err()
	if err != nil {
		fmt.Printf("set score failed, err:%v\n", err)
		return
	}
	val, err := redisdb.Get("score").Result()
	if err != nil {
		fmt.Printf("get score failed, err:%v\n", err)
		return
	}
	fmt.Println("score", val)
	// 不存在的key
	val2, err := redisdb.Get("name").Result()
	if err == redis.Nil {
		fmt.Println("name does not exist")
	} else if err != nil {
		fmt.Printf("get name failed, err:%v\n", err)
		return
	} else {
		fmt.Println("name", val2)
	}
}

// 哈希操作
func redishash() {
	fmt.Println("hash  操作==========================")
	article := Article{"222", "3333333", 10, 0}
	articleKey := "article16"
	redisdb.HMSet(articleKey, ToStringDictionary(&article))
	mapOut := redisdb.HGetAll(articleKey).Val()
	for inx, item := range mapOut {
		fmt.Printf("\n %s:%s", inx, item)
	}
	fmt.Print("\n")
	redisdb.HSet(articleKey, "content", "测试内容")
	mapOut = redisdb.HGetAll(articleKey).Val()
	for inx, item := range mapOut {
		fmt.Printf("\n %s:%s", inx, item)
	}
	fmt.Print("\n")
	view, err := redisdb.HIncrBy(articleKey, "Views", 1).Result()
	if err != nil {
		fmt.Printf("\n HIncrBy error=%s ", err)
	} else {
		fmt.Printf("\n HIncrBy Views=%d ", view)
	}
	fmt.Print("\n")

	mapOut = redisdb.HGetAll(articleKey).Val()
	for inx, item := range mapOut {
		fmt.Printf("\n %s:%s", inx, item)
	}
	fmt.Print("\n")
	//redisdb.HMSet("hash_test","name","nieweibo","age","28","height","dsaf")
	mapOuts := redisdb.HGetAll("hash_test").Val()
	for inx, item := range mapOuts {
		fmt.Printf("%s:%s", inx, item)
	}

}

type Article struct {
	Title      string
	Content    string
	Views      int
	Favourites int
}

func ToStringDictionary(m *Article) map[string]interface{} {
	ArtMap := make(map[string]interface{}, 0)
	ArtMap["Title"] = m.Title
	ArtMap["Content"] = m.Content
	ArtMap["Views"] = m.Views
	ArtMap["Favourites"] = m.Favourites
	return ArtMap
}

// 列表操作
func redislist() {
	fmt.Println("-----------------------welcome to ListDemo-----------------------")
	articleKey := "article"
	result, err := redisdb.RPush(articleKey, "a", "b", "c").Result() //在名称为 key 的list尾添加一个值为value的元素
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("result:", result)

	result, err = redisdb.LPush(articleKey, "d").Result() //在名称为 key 的list头添加一个值为value的元素
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("result:", result)

	length, err := redisdb.LLen(articleKey).Result()
	if err != nil {
		fmt.Println("ListDemo LLen is nil")
	}
	fmt.Println("length: ", length) // 长度

	mapOut, err1 := redisdb.LRange(articleKey, 0, 100).Result()
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	for inx, item := range mapOut {
		fmt.Printf("\n %d:%s \n", inx, item)
	}
}

func GetRedisClientPool() *redis.Client {
	redisdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		PoolSize: 5})

	pong, err := redisdb.Ping().Result()
	if err != nil {
		fmt.Println(pong, err)
	}
	return redisdb
}

// 连接池测试
func connectPoolTest() {
	fmt.Println("-----------------------welcome to connect Pool Test-----------------------")
	client := GetRedisClientPool()
	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100000; j++ {
				client.Set(fmt.Sprintf("name%d", j), fmt.Sprintf("xys%d", j), 0).Err()
				client.Get(fmt.Sprintf("name%d", j)).Result()
			}

			fmt.Printf("PoolStats, TotalConns: %d, IdleConns: %d\n", client.PoolStats().TotalConns, client.PoolStats().IdleConns)
		}()
	}

	wg.Wait()
}

func main() {
	initClient()
	redisstring()
	redislist()
	redishash()
	// redisExample2()
	connectPoolTest()

}
