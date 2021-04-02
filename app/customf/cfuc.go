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

type handle func(name interface{}) bool

var HandleMap map[string]handle

func init() {
	HandleMap = make(map[string]handle)
	HandleMap["CheckIP"] = CheckIP

}

func CheckIP(ip interface{}) bool {
	if ip.(string) == "172.16.17.11" {
		return true
	}
	return false
}
