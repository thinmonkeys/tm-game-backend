package main

import (
	"../api"

	"github.com/aws/aws-lambda-go/lambda"
	chiadaptor "github.com/awslabs/aws-lambda-go-api-proxy/chi"
)

func main() {
	r, err := api.New()
	if err != nil {
		panic(err)
	}
	adapter := chiadaptor.New(r)
	lambda.Start(adapter.ProxyWithContext)
}
