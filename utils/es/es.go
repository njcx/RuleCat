package es

import (
	"errors"
	"fmt"
	log2 "rulecat/utils/log"
	"time"
)

const (
	ESIndex = "index"
)

type UnsupportedVersion struct{}

func (UnsupportedVersion) Error() string {
	return "Unsupported ElasticSearch Client Version"
}

type elasticWrapper interface {
	IndexExists(indices ...string) (bool, error)
	CreateIndex(name string) (bool, error)
	AddAlias(index string, alias string) (bool, error)
	HasAlias(index string, alias string) (bool, error)
	AddBulkReq(index, typeName string, data interface{}) error
	FlushBulk() error
	AddBodyJson(index, typeName string, data interface{}) error
	AddBodyString(index, typeName string, data string) error
}

type ElasticConfig struct {
	Url        []string
	User       string
	Secret     string
	MaxRetries *int
	Sniff      *bool
}

type ElasticSearchService struct {
	EsClient  elasticWrapper
	baseIndex string
}

func (esSvc *ElasticSearchService) Index(typeName string, namespace string) (string, error) {
	date := time.Now()
	dateStr := date.Format("2006.01.02")

	indexName := fmt.Sprintf("%s-%s", esSvc.baseIndex, dateStr)
	if len(namespace) > 0 {
		indexName = fmt.Sprintf("%s-%s", namespace, dateStr)
	}

	if typeName == "" {
		return indexName, errors.New("TypeName  nil")
	}
	// Use the IndexExists service to check if a specified index exists.
	exists, err := esSvc.EsClient.IndexExists(indexName)
	if err != nil {
		return indexName, err
	}
	if !exists {
		ack, err := esSvc.EsClient.CreateIndex(indexName)
		if err != nil {
			return indexName, err
		}
		if !ack {
			return indexName, errors.New("Failed to acknoledge index creation")
		}
	}

	aliasName := esSvc.IndexAlias(typeName)
	hasAlias, err := esSvc.EsClient.HasAlias(indexName, aliasName)
	if err != nil {
		return indexName, err
	}

	if !hasAlias {
		ack, err := esSvc.EsClient.AddAlias(indexName, esSvc.IndexAlias(typeName))
		if err != nil {
			return indexName, err
		}

		if !ack {
			return indexName, errors.New("Failed to acknoledge index alias creation")
		}
	}

	return indexName, nil
}

func (esSvc *ElasticSearchService) IndexAlias(typeName string) string {
	return fmt.Sprintf("%s-%s", esSvc.baseIndex, typeName)
}

func (esSvc *ElasticSearchService) FlushData() error {
	return esSvc.EsClient.FlushBulk()
}

// SaveDataIntoES save metrics and events to ES by using ES client
func (esSvc *ElasticSearchService) SaveData(typeName string, namespace string, sinkData []interface{}) error {
	indexName, err := esSvc.Index(typeName, namespace)
	if err != nil {
		log2.Error.Println(err)
		return err
	}
	for _, data := range sinkData {
		esSvc.EsClient.AddBulkReq(indexName, typeName, data)
	}
	return nil
}

func (esSvc *ElasticSearchService) AddBodyJson(typeName, namespace string, sinkData interface{}) error {
	indexName, err := esSvc.Index(typeName, namespace)
	if err != nil {
		log2.Error.Println(err)
		return err
	}
	return esSvc.EsClient.AddBodyJson(indexName, typeName, sinkData)
}

func (esSvc *ElasticSearchService) AddBodyString(typeName, namespace string, sinkData string) error {
	indexName, err := esSvc.Index(typeName, namespace)
	if err != nil {
		log2.Error.Println(err)
		return err
	}
	return esSvc.EsClient.AddBodyString(indexName, typeName, sinkData)
}

func CreateElasticSearchService(config ElasticConfig, version int) (*ElasticSearchService, error) {
	var esSvc ElasticSearchService
	esSvc.baseIndex = ESIndex
	var err error
	bulkWorkers := 5
	pipeline := ""

	switch version {
	case 6:
		esSvc.EsClient, err = NewEsClient6(config, bulkWorkers, pipeline)
	case 7:
		esSvc.EsClient, err = NewEsClient7(config, bulkWorkers, pipeline)
	default:
		return nil, UnsupportedVersion{}
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to create ElasticSearch client: %v", err)
	}

	return &esSvc, nil
}
