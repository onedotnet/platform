package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchCache implements the Cache interface using Elasticsearch
type ElasticsearchCache struct {
	client    *elasticsearch.Client
	indexName string
}

// NewElasticsearchCache creates a new Elasticsearch cache instance
func NewElasticsearchCache(addresses []string, username, password, indexName string) (*ElasticsearchCache, error) {
	config := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}

	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}

	es := &ElasticsearchCache{
		client:    client,
		indexName: indexName,
	}

	// Create index if it doesn't exist
	err = es.createIndex(context.Background())
	if err != nil {
		return nil, err
	}

	return es, nil
}

// createIndex creates the Elasticsearch index if it doesn't exist
func (e *ElasticsearchCache) createIndex(ctx context.Context) error {
	res, err := e.client.Indices.Exists([]string{e.indexName})
	if err != nil {
		return err
	}

	if res.StatusCode == 200 {
		// Index already exists
		return nil
	}

	// Create index
	mapping := `{
		"mappings": {
			"properties": {
				"key": { "type": "keyword" },
				"value": { "type": "text", "index": false },
				"expiration": { "type": "date" }
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: e.indexName,
		Body:  bytes.NewReader([]byte(mapping)),
	}

	res, err = req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

// Set stores a value in Elasticsearch with the given key and expiration
func (e *ElasticsearchCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Marshal the value
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Create document
	var expirationTime *time.Time
	if expiration > 0 {
		t := time.Now().Add(expiration)
		expirationTime = &t
	}

	doc := map[string]interface{}{
		"key":        key,
		"value":      string(data),
		"expiration": expirationTime,
	}

	// Marshal document
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	// Index document
	req := esapi.IndexRequest{
		Index:      e.indexName,
		DocumentID: key,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

// Get retrieves a value from Elasticsearch using the given key
func (e *ElasticsearchCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Check if document exists and is not expired
	req := esapi.GetRequest{
		Index:      e.indexName,
		DocumentID: key,
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return fmt.Errorf("key %s not found", key)
	}

	if res.IsError() {
		return fmt.Errorf("error getting document: %s", res.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}

	source, ok := response["_source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response format")
	}

	// Check expiration
	if expStr, ok := source["expiration"].(string); ok && expStr != "" {
		expTime, err := time.Parse(time.RFC3339, expStr)
		if err != nil {
			return err
		}

		if time.Now().After(expTime) {
			// Document is expired, delete it
			e.Delete(ctx, key)
			return fmt.Errorf("key %s has expired", key)
		}
	}

	// Unmarshal value
	valueStr, ok := source["value"].(string)
	if !ok {
		return fmt.Errorf("invalid value format")
	}

	return json.Unmarshal([]byte(valueStr), dest)
}

// Delete removes an item from Elasticsearch using the given key
func (e *ElasticsearchCache) Delete(ctx context.Context, key string) error {
	req := esapi.DeleteRequest{
		Index:      e.indexName,
		DocumentID: key,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 404 is ok for delete
	if res.StatusCode != 200 && res.StatusCode != 404 {
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	return nil
}

// Clear removes all items from the Elasticsearch index
func (e *ElasticsearchCache) Clear(ctx context.Context) error {
	// Delete and recreate the index
	req := esapi.IndicesDeleteRequest{
		Index: []string{e.indexName},
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 404 is ok for delete
	if res.StatusCode != 200 && res.StatusCode != 404 {
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	// Create new index
	return e.createIndex(ctx)
}

// Close closes the Elasticsearch connection (no-op as Elasticsearch client doesn't require explicit closing)
func (e *ElasticsearchCache) Close() error {
	return nil
}
