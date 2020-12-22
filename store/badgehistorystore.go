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

//BadgeHistoryStore creates a new store for BadgeHistory instances.
func NewBadgeHistoryStore(region, tableName string) (cs BadgeHistoryStore, err error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}
	cfg.Region = region

	cs.Client = dynamodb.New(cfg)
	cs.TableName = aws.String(tableName)
	return
}

func DefaulBadgeHistoryStore() (cs BadgeHistoryStore, err error) {
	return NewScoreHistoryStore("eu-west-1", "UserBadgeHistory")
}

// BadgeHistoryStore stores user's BadgeHistory records in DynamoDB.
type BadgeHistoryStore struct {
	Client    dynamodbiface.ClientAPI
	TableName *string
}

// BadgeHistoryRecord is the data used to store challenges.
type BadgeHistoryRecord struct {
	BadgeCode   	string `json:"BadgeCode"`
	CustomerCIF     string `json:"CustomerCIF"`
	DateAwarded  time.Time `json:"DateAwarded"`
}

const badgeRecordName = "badges"

// Put the record in DynamoDB.
func (store BadgeHistoryStore) Put(record ScoreHistoryRecord) (err error) {
	item, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return
	}
	item["CIFWithBadgeCode"] = dynamodb.AttributeValue{
		S: aws.String(record.CustomerCIF + record.BadgeCode)
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
func (store BadgeHistoryStore) Get(cif string) (record []BadgeHistoryRecord, err error) {

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
