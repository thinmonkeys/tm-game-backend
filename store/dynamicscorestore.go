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

// DynamicScoreStore creates a new store for ChallengeRecord instances.
func NewDynamicScoreStore(region, tableName string) (cs DynamicScoreStore, err error) {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return
	}
	cfg.Region = region

	cs.Client = dynamodb.New(cfg)
	cs.TableName = aws.String(tableName)
	return
}

func DefaultDynamicScoreStore() (cs DynamicScoreStore, err error) {
	return NewDynamicScoreStore("eu-west-1", "UserScoreDataTable")
}

// DynamicScoreStore stores Customer Score records in DynamoDB.
type DynamicScoreStore struct {
	Client    dynamodbiface.ClientAPI
	TableName *string
}

// DynamicScoreRecord is the data used to store challenges.
type DynamicScoreRecord struct {
	CustomerCIF  string `json:"customerCif"`
	Score    	int `json:"score"`
	LastUpdatedContactDetails    time.Time `json:"lastUpdatedContactDetails"`
	LastUpdatedStandingOrders 	 time.Time `json:"lastUpdatedStandingOrders"`
	LastUpdatedDirectDebits 	 time.Time `json:"lastUpdatedDirectDebits"`
	LastUpdatedIncomes 	 		 time.Time `json:"lastUpdatedIncomes"`
}

const scoreRecordName = "score"

// Put the record in DynamoDB.
func (store DynamicScoreStore) Put(vr DynamicScoreRecord) (err error) {
	item, err := dynamodbattribute.MarshalMap(vr)
	if err != nil {
		return
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
func (store DynamicScoreStore) Get(cif string) (record DynamicScoreRecord, ok bool, err error) {
	input := &dynamodb.GetItemInput{
		ConsistentRead:   aws.Bool(true),
		Key: map[string]dynamodb.AttributeValue{
			"CustomerCIF": {
				S: aws.String(cif),
			},
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
	if err != nil {
		return
	}
	ok = record.CustomerCIF != ""
	return
}