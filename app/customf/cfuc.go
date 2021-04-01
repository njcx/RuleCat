package customf

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

type handle func(name string) bool

var HandleMap map[string]handle

func init() {
	HandleMap = make(map[string]handle)
	HandleMap["CheckIP"] = CheckIP

}

func CheckIP(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}
	return false
}
