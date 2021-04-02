package app

import (
	"github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"rulecat/app/customf"
	"rulecat/utils"
	"rulecat/utils/cache"
	"rulecat/utils/json"
	log2 "rulecat/utils/log"
	"rulecat/utils/workerpool"
	"strings"
	"sync"
	"time"
)

var Json1 = jsoniter.ConfigCompatibleWithStandardLibrary
var Tc = cache.New(60*60*time.Second, 1*time.Second)
var RuleList []map[interface{}]interface{}

type engine struct {
	RuleType string
	InPutC   chan string
	OutPutC  chan *sync.Map
	Json     string
}

func NewEngine(ruleType string) *engine {
	return &engine{RuleType: ruleType, InPutC: make(chan string, 48), OutPutC: make(chan *sync.Map, 48)}
}

func (e *engine) ReadRules() {
	var rulesPath []string
	rulesListPath, err := utils.GetAllFile(utils.GetCurrentPath()+"/etc/rules/"+e.RuleType, rulesPath)
	if err != nil {
		log2.Error.Fatalf("Get rule file dir err %s  %v ", rulesListPath, err)
	}
	for _, ruleFile := range rulesListPath {
		rule, err := ioutil.ReadFile(ruleFile)

		if err != nil {
			log2.Error.Fatalf("Get rule file err  %s %v ", rule, err)
		}
		m := make(map[interface{}]interface{})
		err = yaml.Unmarshal(rule, &m)
		if err != nil {
			log2.Error.Fatalf("Unmarshal rule file err  %s %v ", rule, err)
		}
		RuleList = append(RuleList, m)
	}
	log2.Info.Printf("Rule file have been load %s ", e.RuleType)

}

func (e *engine) ResCheck(threadNum int, outPut chan *sync.Map) {
	p := workerpool.NewWorkerPool(threadNum)
	p.Run()
	go func() {
		for {
			sc := &engine{Json: <-e.InPutC, OutPutC: outPut}
			p.JobQueue <- sc

		}
	}()
}

func (e *engine) Do() error {
	RuleCheckFuc(e.Json, RuleList, e.OutPutC)
	return nil

}

func RuleCheckFuc(s string, r []map[interface{}]interface{}, outPut chan *sync.Map) {
	for _, rule := range r {
		if rule["state"] == "enable" {
			var detectList int
			for _, detectItem := range rule["detect_list"].([]interface{}) {
				if detectItem.(map[interface{}]interface{})["type"] == "equal" {
					if json.Get(s, detectItem.(map[interface{}]interface{})["field"].(string)).String() ==
						detectItem.(map[interface{}]interface{})["rule"].(string) {
						detectList++
					}
				}
				if detectItem.(map[interface{}]interface{})["type"] == "re" {
					match, _ := regexp.MatchString(detectItem.(map[interface{}]interface{})["rule"].(string),
						json.Get(s, detectItem.(map[interface{}]interface{})["field"].(string)).String())
					if match {
						detectList++
					}
				}
				if detectItem.(map[interface{}]interface{})["type"] == "in" {
					if strings.Contains(json.Get(s, detectItem.(map[interface{}]interface{})["field"].(string)).String(),
						detectItem.(map[interface{}]interface{})["rule"].(string)) {
						detectList++
					}
				}

				if detectItem.(map[interface{}]interface{})["type"] == "customf" {
					handelf := customf.HandleMap[detectItem.(map[interface{}]interface{})["rule"].(string)]
					if handelf(json.Get(s, detectItem.(map[interface{}]interface{})["field"].(string)).String()) {
						detectList++
					}
				}
			}
			if rule["rule_type"] == "and" {
				if len(rule["detect_list"].([]interface{})) == detectList {
					var tmpMap map[string]interface{}
					_ = Json1.Unmarshal([]byte(s), &tmpMap)
					sMap, _ := utils.MapToSMap(rule)
					sMap.Store("data", tmpMap)
					outPut <- sMap
				}
			}
			if rule["rule_type"] == "or" {
				if detectList > 0 {
					var tmpMap map[string]interface{}
					_ = Json1.Unmarshal([]byte(s), &tmpMap)
					sMap, _ := utils.MapToSMap(rule)
					sMap.Store("data", tmpMap)
					outPut <- sMap
				}
			}
			if rule["rule_type"] == "frequency_and" {
				if len(rule["detect_list"].([]interface{})) == detectList {
					value, found := Tc.Get(json.Get(s, rule["key"].(string)).String())
					if found {
						if value.(int) >= rule["time_interval"].(map[interface{}]interface{})["times"].(int) {
							var tmpMap map[string]interface{}
							_ = Json1.Unmarshal([]byte(s), &tmpMap)
							sMap, _ := utils.MapToSMap(rule)
							sMap.Store("data", tmpMap)
							outPut <- sMap

						} else {
							_ = Tc.Increment(json.Get(s, rule["key"].(string)).String(), 1)
						}
					} else {
						second := rule["time_interval"].(map[interface{}]interface{})["second"].(int)
						Tc.Set(json.Get(s, rule["key"].(string)).String(), 1, time.Duration(second)*time.Second)
					}
				}
			}
			if rule["rule_type"] == "frequency_or" {
				if detectList > 0 {
					value, found := Tc.Get(json.Get(s, rule["key"].(string)).String())
					if found {
						if value.(int) >= rule["time_interval"].(map[interface{}]interface{})["times"].(int) {
							var tmpMap map[string]interface{}
							_ = Json1.Unmarshal([]byte(s), &tmpMap)
							sMap, _ := utils.MapToSMap(rule)
							sMap.Store("data", tmpMap)
							outPut <- sMap
						} else {
							_ = Tc.Increment(json.Get(s, rule["key"].(string)).String(), 1)
						}
					} else {
						second := rule["time_interval"].(map[interface{}]interface{})["second"].(int)
						Tc.Set(json.Get(s, rule["key"].(string)).String(), 1, time.Duration(second)*time.Second)
					}
				}
			}
		}
	}
}
