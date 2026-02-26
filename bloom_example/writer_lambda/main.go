// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"writer_lambda/client"
	"writer_lambda/config"

	"github.com/aws/aws-lambda-go/lambda"
)

// Request represents the incoming lambda request
type Request struct {
	Key  string   `json:"key,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

func writerHandler(ctx context.Context, request Request) (client.Response, error) {
	conf, err := config.LoadConf()
	if err != nil {
		return errResponse(err), nil
	}

	if conf.DryRun {
		return client.Response{
			StatusCode: 200,
			Headers:    client.Headers,
			Body:       "dry run test",
		}, nil
	}

	c, err := client.NewClient(
		client.AsService("writer_lambda"),
		client.WithHTTPHost(conf.CacheHost),
	)
	if err != nil {
		return errResponse(err), nil
	}

	var resp client.Response

	// Handle single key or multiple keys
	if request.Key != "" {
		resp, err = c.AddToBloom(ctx, request.Key)
	} else if len(request.Keys) > 0 {
		resp, err = c.AddMultipleToBloom(ctx, request.Keys)
	} else {
		return errResponse(fmt.Errorf("no key or keys provided")), nil
	}

	if err != nil {
		return errResponse(err), nil
	}

	// log success
	log.Print(fmt.Sprintf("status code: %d, body: %s", resp.StatusCode, resp.Body))

	return resp, nil
}

func main() {
	log.SetOutput(os.Stdout)

	lambda.Start(writerHandler)
}

func errResponse(err error) client.Response {
	errBody, _ := json.Marshal(map[string]string{"error": err.Error()})
	return client.Response{
		StatusCode: 500,
		Headers:    client.Headers,
		Body:       string(errBody),
	}
}
