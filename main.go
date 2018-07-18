package main

import (
	"context"
	"encoding/xml"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/pruser/allegro-feed-generator/config"
	"github.com/pruser/allegro-feed-generator/request"
)

type StringMapWrapper map[string]string

func (w StringMapWrapper) Get(key string) string {
	return w[key]
}

type ErrorMessage struct {
	XMLName    xml.Name `xml:"Error"`
	Message    string   `xml:"Message"`
	StatusCode int      `xml:"StatusCode"`
}

func createAtomResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return createResponse("application/atom+xml", statusCode, body)
}

func createResponse(contentType string, statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    map[string]string{"content-type": contentType},
	}
}

func createErrorResponse(statusCode int, err error) (events.APIGatewayProxyResponse, error) {
	errMsg := ErrorMessage{Message: err.Error(), StatusCode: statusCode}
	body, err := xml.Marshal(errMsg)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	return createResponse("application/xml", statusCode, string(body)), nil
}

func main() {
	config, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	handler := request.NewRequestHandler(config.WebAPIKey, config.WebAPIUrl, config.UrlBase)
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		res, erro := handler.CreateFeedImpl(StringMapWrapper(request.QueryStringParameters))
		if erro != nil {
			response, err := createErrorResponse(erro.StatusCode, erro)
			return response, err
		}
		return createAtomResponse(http.StatusOK, res), nil
	})
}
