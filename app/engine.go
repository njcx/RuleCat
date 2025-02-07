package app

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"rulecat/utils"
	"rulecat/utils/cache"
	"rulecat/utils/json"
	log2 "rulecat/utils/log"
	"rulecat/utils/workerpool"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Json1 = jsoniter.ConfigCompatibleWithStandardLibrary
var Tc = cache.New(60*60*time.Second, 1*time.Second)
var RuleList []map[string]interface{}

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
	rulesListPath, err := utils.GetAllFile(utils.GetCurrentPath()+"/etc/"+e.RuleType+"_rules", rulesPath)
	if err != nil {
		log2.Error.Fatalf("Get rule file dir err %s  %v ", rulesListPath, err)
	}
	for _, ruleFile := range rulesListPath {
		rule, err := os.ReadFile(ruleFile)
		if err != nil {
			log2.Error.Fatalf("Get rule file err  %s %v ", rule, err)
		}
		m := make(map[string]interface{})
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

func RuleCheckFuc(s string, r []map[string]interface{}, outPut chan *sync.Map) {
	for _, rule := range r {
		if rule["state"] != "enable" {
			continue
		}
		detectList, infoMap := checkDetectList(s, rule)
		handleRuleType(s, rule, detectList, infoMap, outPut)
	}
}

func interfaceToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func isArray(v interface{}) bool {
	switch v.(type) {
	case []int, []string, []float64:
		return true
	default:
		return false
	}
}

func convertToStringSlice(v interface{}) ([]string, error) {
	switch v := v.(type) {
	case []int:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = strconv.Itoa(val)
		}
		return result, nil
	case []string:
		return v, nil
	case []float64:
		result := make([]string, len(v))
		for i, val := range v {
			result[i] = strconv.FormatFloat(val, 'f', -1, 64)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func isInList(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

func ConvertMap(inputMap map[interface{}]interface{}) (map[string]interface{}, error) {
	convertedMap := make(map[string]interface{})
	for k, v := range inputMap {
		keyStr, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("key %v is not a string", k)
		}
		switch vTyped := v.(type) {
		case map[interface{}]interface{}:
			nestedMap, err := ConvertMap(vTyped)
			if err != nil {
				return nil, err
			}
			convertedMap[keyStr] = nestedMap
		case []interface{}:
			convertedSlice := make([]interface{}, len(vTyped))
			for i, item := range vTyped {
				if subMap, ok := item.(map[interface{}]interface{}); ok {
					nestedMap, err := ConvertMap(subMap)
					if err != nil {
						return nil, err
					}
					convertedSlice[i] = nestedMap
				} else {
					convertedSlice[i] = item
				}
			}
			convertedMap[keyStr] = convertedSlice
		default:
			convertedMap[keyStr] = vTyped
		}
	}
	return convertedMap, nil
}

func convertToMapSlice(slice []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(slice))
	for _, item := range slice {
		if m, ok := item.(map[interface{}]interface{}); ok {
			convertedMap, err := ConvertMap(m)
			if err != nil {
				fmt.Printf("Skipping invalid element: %v (type: %T) due to error: %v\n", item, item, err)
				continue
			}
			result = append(result, convertedMap)
		} else {
			fmt.Printf("Skipping invalid element: %v (type: %T)\n", item, item)
		}
	}
	return result
}

func checkDetectList(s string, rule map[string]interface{}) (int, map[string]string) {
	detectList := 0
	infoMap := make(map[string]string)
	for _, detectItem := range convertToMapSlice(rule["detect_list"].([]interface{})) {
		field := detectItem["field"].(string)
		value := json.Get(s, field).String()
		switch detectItem["type"].(string) {
		case "equal":
			if value == interfaceToString(detectItem["rule"]) {
				detectList++
			}
		case "re":
			pattern := detectItem["rule"].(string)
			if detectItem["ignore-case"].(bool) {
				match, _ := regexp.MatchString(`(?i)`+pattern, value)
				if match {
					detectList++
				}
			} else {
				match, _ := regexp.MatchString(pattern, value)
				if match {
					detectList++
				}
			}
		case "in":
			if isArray(detectItem["rule"]) {
				listTmp, err := convertToStringSlice(detectItem["rule"])
				if err != nil {
					continue
				}
				if isInList(listTmp, value) {
					detectList++
				}
			} else {
				if strings.Contains(value, detectItem["rule"].(string)) {
					detectList++
				}
			}
		case "customf":
			handelFuc := HandleMap[detectItem["rule"].(string)]
			successHit, tmpMap := handelFuc(value)
			if successHit {
				detectList++
			}
			infoMap = tmpMap
		}
	}
	return detectList, infoMap
}

func handleRuleType(s string, rule map[string]interface{}, detectList int, infoMap map[string]string, outPut chan *sync.Map) {
	ruleType := rule["rule_type"].(string)
	detectListCount := len(rule["detect_list"].([]interface{}))
	switch ruleType {
	case "and":
		if detectListCount == detectList {
			sendResult(s, rule, infoMap, outPut)
		}
	case "or":
		if detectList > 0 {
			sendResult(s, rule, infoMap, outPut)
		}
	case "frequency_and", "frequency_or":
		if detectListCount == detectList || (ruleType == "frequency_or" && detectList > 0) {
			key := rule["key"].(string)
			handleFrequency(key, rule, s, infoMap, outPut)
		}
	}
}

func sendResult(s string, rule map[string]interface{}, infoMap map[string]string, outPut chan *sync.Map) {
	var tmpMap map[string]interface{}
	_ = Json1.Unmarshal([]byte(s), &tmpMap)
	sMap, _ := utils.MapToSMap(rule)
	sMap.Store("data", tmpMap)
	sMap.Store("extra_message", infoMap)
	outPut <- sMap
}

func handleFrequency(key string, rule map[string]interface{}, s string, infoMap map[string]string, outPut chan *sync.Map) {
	FrequencyMap, _ := ConvertMap(rule["time_interval"].(map[interface{}]interface{}))
	times := FrequencyMap["times"].(int)
	second := FrequencyMap["second"].(int)
	value, found := Tc.Get(key)
	if found {
		if value.(int) >= times {
			sendResult(s, rule, infoMap, outPut)
		} else {
			Tc.Increment(key, 1)
		}
	} else {
		Tc.Set(key, 1, time.Duration(second)*time.Second)
	}
}
