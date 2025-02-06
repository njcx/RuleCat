package utils

import (
	"bytes"
	Json "encoding/json"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetCurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func GetAllFile(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = GetAllFile(fullDir, s)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

func MarshalSMapToJSON(m *sync.Map) ([]byte, error) {
	tmpMap := make(map[interface{}]interface{})
	m.Range(func(k, v interface{}) bool {
		tmpMap[k] = v
		return true
	})
	return json.Marshal(tmpMap)
}

func UnmarshalJSONToSMap(data []byte) (*sync.Map, error) {
	var tmpMap map[interface{}]interface{}
	m := &sync.Map{}

	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return m, err
	}

	for key, value := range tmpMap {
		m.Store(key, value)
	}
	return m, nil
}

func UnmarshalJSONToMap(data []byte) (map[interface{}]interface{}, error) {
	var tmpMap map[interface{}]interface{}

	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return nil, err
	}

	return tmpMap, nil
}

func MapToSMap(tmpMap map[interface{}]interface{}) (*sync.Map, error) {

	m := &sync.Map{}

	for key, value := range tmpMap {
		m.Store(key, value)
	}
	return m, nil
}

func SMapToMap(tmpMap *sync.Map) map[interface{}]interface{} {

	var m = make(map[interface{}]interface{})
	tmpMap.Range(func(k, v interface{}) bool {
		m[k] = v
		return true
	})
	return m
}

func GetConfig(configFile []byte) (result map[string]interface{}, err error) {
	err = yaml.Unmarshal(configFile, &result)
	return result, err
}

func FormatJson(data []byte) string {
	var out bytes.Buffer
	Json.Indent(&out, data, "", "    ")
	return out.String()
}

func WriteFile(path string, str string) {
	_, b := IsFile(path)
	var f *os.File
	var err error
	if b {
		f, _ = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		f, err = os.Create(path)
	}

	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Println("err = ", err)
		}
	}()

	if err != nil {
		fmt.Println("err = ", err)
		return
	}
	_, err = f.WriteString(str)
	if err != nil {
		fmt.Println("err = ", err)
	}
}

func IsExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}

func IsDir(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && f.IsDir()
}

func IsFile(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && !f.IsDir()
}

func ServerBanner() {
	template := `
{{ .Title "RuleCat" "" 4 }}
GoVersion: {{ .GoVersion }}
GOOS: {{ .GOOS }}
GOARCH: {{ .GOARCH }}
NumCPU: {{ .NumCPU }}
Now: {{ .Now "2006-01-02 15:04:05" }}
	
`
	banner.InitString(os.Stdout, true, true, template)
}
