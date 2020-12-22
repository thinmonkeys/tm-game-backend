package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbiface"
)

// ScoreHistoryStore creates a new store for HistoryScore instances.
func NewScoreHistoryStore(region, tableName string) (cs ScoreHistoryStore, err error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}
	cfg.Region = region

	cs.Client = dynamodb.New(cfg)
	cs.TableName = aws.String(tableName)
	return
}

func DefaultScoreHistoryStore() (cs ScoreHistoryStore, err error) {
	return NewScoreHistoryStore("eu-west-1", "UserScoreHistory")
}

// ScoreHistoryStore stores user's ScoreHistory records in DynamoDB.
type ScoreHistoryStore struct {
	Client    dynamodbiface.ClientAPI
	TableName *string
}

// DynamicScoreRecord is the data used to store challenges.
type ScoreHistoryRecord struct {
	CategoryCode   string    `json:"CategoryCode"`
	CustomerCIF    string    `json:"CustomerCIF"`
	LastConfirmed  time.Time `json:"LastConfirmed"`
	LastScored     time.Time `json:"LastScored"`
	TimesConfirmed int       `json:"TimesConfirmed"`
	TimesScored    int       `json:"TimesScored"`
}

// Put the record in DynamoDB.
func (store ScoreHistoryStore) Put(record ScoreHistoryRecord) (err error) {
	item, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return
	}
	item["CIFWithCategory"] = getKeyAttribute(record.CustomerCIF, record.CategoryCode)
	pir := store.Client.PutItemRequest(&dynamodb.PutItemInput{
		TableName: store.TableName,
		Item:      item,
	})
	_, err = pir.Send(context.Background())

	if err != nil {
		return
	}
	return
}

// Get retrieves data from DynamoDB.
func (store ScoreHistoryStore) GetAll(cif string) (records []ScoreHistoryRecord, err error) {
	input := &dynamodb.ScanInput{
		ConsistentRead:   aws.Bool(true),
		FilterExpression: aws.String("CustomerCIF = :cif"),
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":cif": {
				S: aws.String(cif),
			},
		},
		TableName: store.TableName,
	}
	getReq := store.Client.ScanRequest(input)

	getResult, err := getReq.Send(context.Background())

	if err != nil {
		return
	}
	err = dynamodbattribute.UnmarshalListOfMaps(getResult.Items, &records)
	return
}

func (store ScoreHistoryStore) Get(cif string, categoryCode string) (record ScoreHistoryRecord, ok bool, err error) {
	input := &dynamodb.GetItemInput{
		ConsistentRead:   aws.Bool(true),
		Key: map[string]dynamodb.AttributeValue{
			"CIFWithCategory": getKeyAttribute(cif, categoryCode),
		},
		TableName: store.TableName,
	}
 	getReq := store.Client.GetItemRequest(input)

	getResult, err := getReq.Send(context.Background())

	if err != nil {
		return
	}
	if getResult.Item == nil {
		ok = false
		return
	}
	err = dynamodbattribute.UnmarshalMap(getResult.Item, &record)
	ok = (err == nil && record.CustomerCIF == cif)
	return
}

func getKeyAttribute(cif string, categoryCode string) dynamodb.AttributeValue {
	return dynamodb.AttributeValue{
		S: aws.String(cif + categoryCode),
	}
}