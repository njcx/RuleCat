package utils

import (
	"bytes"
	Json "encoding/json"
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/json-iterator/go"
	"os"
	log2 "rulecat/utils/log"
	"strings"
	"sync"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetCurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log2.Error.Fatalln(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func GetAllFile(pathname string, s []string) ([]string, error) {
	rd, err := os.ReadDir(pathname)
	if err != nil {
		log2.Error.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = GetAllFile(fullDir, s)
			if err != nil {
				log2.Error.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

func ConvertToStringMap(m interface{}) interface{} {
	switch x := m.(type) {
	case map[interface{}]interface{}:
		newMap := map[string]interface{}{}
		for k, v := range x {
			newMap[fmt.Sprint(k)] = ConvertToStringMap(v)
		}
		return newMap
	case map[string]interface{}:
		newMap := map[string]interface{}{}
		for k, v := range x {
			newMap[k] = ConvertToStringMap(v)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(x))
		for i, v := range x {
			newSlice[i] = ConvertToStringMap(v)
		}
		return newSlice
	default:
		return x
	}
}

func MarshalSMapToJSON(m *sync.Map) ([]byte, error) {

	if m == nil {
		return nil, fmt.Errorf("sync.Map is nil")
	}

	tmpMap := make(map[interface{}]interface{})
	m.Range(func(k, v interface{}) bool {
		if k == nil || v == nil {
			return true
		}
		tmpMap[k] = v
		return true
	})

	convertedMap := ConvertToStringMap(tmpMap)
	data, err := Json.Marshal(convertedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return data, nil
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

func FormatJson(data []byte) string {
	var out bytes.Buffer
	Json.Indent(&out, data, "", "    ")
	return out.String()
}

func WriteFile(path string, str string) error {
	_, fileExists := IsFile(path)

	var f *os.File
	var err error

	if fileExists {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		f, err = os.Create(path)
	}

	if err != nil {
		log2.Error.Printf("Failed to open/create file: %v", err)
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log2.Error.Printf("Failed to close file: %v", cerr)
		}
	}()

	_, err = f.WriteString(str)
	if err != nil {
		log2.Error.Printf("Failed to write to file: %v", err)
		return err
	}

	return nil
}

func IsExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
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
