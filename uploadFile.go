package awsPractice

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func uploadFile(filepath string, bucket string, key string) error {
	// Opening the file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %v", err)
	}

	// Creating a new S3 client
	s3Client := s3.NewFromConfig(awsSession)

	// Uploading the file to the s3 bucket
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("Error uploading file: %v", err)
	}

	return nil
}

func notifyLambda(payload map[string]string) error {
	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %v", err)
	}

	// Creating a new lambda client
	lambdaClient := lambda.NewFromConfig(awsSession)

	// Marshalling the payload into bytes
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshalling payload: %v", err)
	}

	// Invoking the lambda function
	_, err = lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(payload["lambdaName"]),
		Payload:      payloadBytes,
	})
	if err != nil {
		return fmt.Errorf("Error invoking lambda function: %v", err)
	}

	return nil
}

func main() {
	// Parsing the bucket, key, lambda, and filepath flags from the command line
	filepath := flag.String("filepath", "", "Path to the file that we want to upload to AWS")
	bucket := flag.String("bucket", "", "Name of the S3 bucket that we want to upload the file to")
	key := flag.String("key", "", "Name of the key that we want to upload the file to")
	lambdaName := flag.String("lambda", "", "Name of the Lambda function that we want to notify")
	flag.Parse()

	// Uploading the file to the s3 bucker using a goroutine
	s3Upload := make(chan error)

	go func() {
		s3Upload <- uploadFile(*filepath, *bucket, *key)
	}()

	err := <-s3Upload
	if err != nil {
		fmt.Println("Error uploading file: %v", err)
		return
	}

	// Notifying a lambda function using a goroutine, and also using AWS SQS to send a message to receivers
	lambdaUpload := make(chan error)

	// Configuring the payload for the lambda function
	payload := map[string]string{
		"bucket":     *bucket,
		"key":        *key,
		"lambdaName": *lambdaName,
	}
	go func() {
		lambdaUpload <- notifyLambda(payload)
	}()

	err = <-lambdaUpload
	if err != nil {
		fmt.Printf("Error notifying lambda function: %v", err)
		return
	}

	fmt.Println("File uploaded and lambda function notified successfully")
}
