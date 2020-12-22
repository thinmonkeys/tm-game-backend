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

func DefaulScoreHistoryStore() (cs DynamicScoreStore, err error) {
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

const scoreRecordName = "history"

// Put the record in DynamoDB.
func (store ScoreHistoryStore) Put(record ScoreHistoryRecord) (err error) {
	item, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return
	}
	item["CIFWithCategory"] = dynamodb.AttributeValue{
		S: aws.String(record.CustomerCIF + record.CategoryCode)
	}
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
func (store ScoreHistoryStore) Get(cif string) (record []ScoreHistoryRecord, err error) {
	input := &dynamodb.ScanInput{
		ConsistentRead:   aws.Bool(true),
		FilterExpression: aws.String("CustomerCIF = CIF"),
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			"CIF": {
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
	err = dynamodbattribute.UnmarshalMap(getResult.Items, &record)
	if err != nil {
		return
	}
}
