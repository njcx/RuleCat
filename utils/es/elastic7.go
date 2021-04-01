package es

import (
	"context"
	"fmt"
	elastic7 "github.com/olivere/elastic/v7"
	"github.com/pborman/uuid"
	"time"
)

type Elastic7Wrapper struct {
	client        *elastic7.Client
	pipeline      string
	bulkProcessor *elastic7.BulkProcessor
}

func NewEsClient7(config ElasticConfig, bulkWorkers int, pipeline string) (*Elastic7Wrapper, error) {
	var startupFns []elastic7.ClientOptionFunc
	if len(config.Url) > 0 {
		startupFns = append(startupFns, elastic7.SetURL(config.Url...))
	}
	if config.User != "" && config.Secret != "" {
		startupFns = append(startupFns, elastic7.SetBasicAuth(config.User, config.Secret))
	}
	if config.MaxRetries != nil {
		startupFns = append(startupFns, elastic7.SetMaxRetries(*config.MaxRetries))
	}
	if config.Sniff != nil {
		startupFns = append(startupFns, elastic7.SetSniff(*config.Sniff))
	}
	client, err := elastic7.NewClient(startupFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to an ElasticSearch Client: %v", err)
	}
	bps, err := client.BulkProcessor().
		Name("ElasticSearchWorker").
		Workers(bulkWorkers).
		After(bulkAfterCBV7).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(10 * time.Second). // commit every 10s
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to an ElasticSearch Bulk Processor: %v", err)
	}

	return &Elastic7Wrapper{client: client, bulkProcessor: bps, pipeline: pipeline}, nil
}

func (es *Elastic7Wrapper) IndexExists(indices ...string) (bool, error) {
	return es.client.IndexExists(indices...).Do(context.Background())
}

func (es *Elastic7Wrapper) CreateIndex(name string) (bool, error) {
	res, err := es.client.CreateIndex(name).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic7Wrapper) getAliases(index string) (*elastic7.AliasesResult, error) {
	return es.client.Aliases().Index(index).Do(context.Background())
}

func (es *Elastic7Wrapper) AddAlias(index string, alias string) (bool, error) {
	res, err := es.client.Alias().Add(index, alias).Do(context.Background())
	if err != nil {
		return false, err
	}
	return res.Acknowledged, err
}

func (es *Elastic7Wrapper) HasAlias(indexName string, aliasName string) (bool, error) {
	aliases, err := es.getAliases(indexName)
	if err != nil {
		return false, err
	}
	return aliases.Indices[indexName].HasAlias(aliasName), nil
}

func (es *Elastic7Wrapper) AddBulkReq(index, typeName string, data interface{}) error {
	req := elastic7.NewBulkIndexRequest().
		Index(index).
		Type(typeName).
		Id(uuid.NewUUID().String()).
		Doc(data)
	if es.pipeline != "" {
		req.Pipeline(es.pipeline)
	}
	es.bulkProcessor.Add(req)
	return nil
}

func (es *Elastic7Wrapper) FlushBulk() error {
	return es.bulkProcessor.Flush()
}

func (es *Elastic7Wrapper) AddBodyJson(index, typeName string, data interface{}) error {
	put1, err := es.client.Index().
		Index(index).
		Type(typeName).
		Id(uuid.NewUUID().String()).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		fmt.Println("Failed to indexed data: %v", err)
		return err
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
	return nil
}

func (es *Elastic7Wrapper) AddBodyString(index, typeName string, data string) error {
	put2, err := es.client.Index().
		Index(index).
		Type(typeName).
		Id(uuid.NewUUID().String()).
		BodyString(data).
		Do(context.Background())
	if err != nil {
		fmt.Println("Failed to indexed data: %v", err)
		return err
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)
	return nil
}

func bulkAfterCBV7(_ int64, _ []elastic7.BulkableRequest, response *elastic7.BulkResponse, err error) {
	if err != nil {
		fmt.Println("Failed to execute bulk operation to ElasticSearch: %v", err)
	}

	if response.Errors {
		for _, list := range response.Items {
			for name, itm := range list {
				if itm.Error != nil {
					fmt.Println("Failed to execute bulk operation to ElasticSearch on %s: %v", name, itm.Error)
				}
			}
		}
	}
}
