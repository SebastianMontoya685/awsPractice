package main

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
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("creating AWS session: %w", err)
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
		return fmt.Errorf("uploading file: %w", err)
	}

	return nil
}

func notifyLambda(payload map[string]string) error {
	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("creating AWS session: %w", err)
	}

	// Creating a new lambda client
	lambdaClient := lambda.NewFromConfig(awsSession)

	// Marshalling the payload into bytes
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	// Invoking the lambda function
	_, err = lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(payload["lambdaName"]),
		Payload:      payloadBytes,
	})
	if err != nil {
		return fmt.Errorf("invoking lambda function: %w", err)
	}

	return nil
}

func getKeys(bucket string) ([]string, error) {
	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %v", err)
	}

	// Definingin the keys that we want to return
	var (
		keys              []string
		continuationToken *string
	)

	// Creating a new S3 client
	s3Client := s3.NewFromConfig(awsSession)

	// Looping through the keys in the S3 Bucket
	for {
		// Listing the keys in the S3 Bucket
		listObjectsOutput, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			// Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error listing objects: %v", err)
		}

		// Collect keys from this page
		for _, object := range listObjectsOutput.Contents {
			keys = append(keys, aws.ToString(object.Key))
		}

		if !aws.ToBool(listObjectsOutput.IsTruncated) {
			break
		}

		continuationToken = listObjectsOutput.NextContinuationToken
	}

	return keys, nil
}

func listLambdaFunctions() error {
	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	// Creating a new lambda client
	lambdaClient := lambda.NewFromConfig(awsSession)

	// Listing the lambda functions that we have access to
	listFunctionsOutput, err := lambdaClient.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		return fmt.Errorf("error listing lambda functions: %v", err)
	}

	// Printing the lambda functions that we have access to
	for _, function := range listFunctionsOutput.Functions {
		fmt.Println("-", aws.ToString(function.FunctionName))
	}

	return nil
}

func listS3Buckets() error {
	// Creating a new AWS session
	awsSession, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	// Creating a new S3 client
	s3Client := s3.NewFromConfig(awsSession)

	// Listing the S3 buckets that we have
	listBucketsOutput, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("error listing S3 buckets: %v", err)
	}

	fmt.Println("S3 buckets:")
	for _, bucket := range listBucketsOutput.Buckets {
		fmt.Println("-", aws.ToString(bucket.Name))
	}

	return nil
}

func main() {
	// Parsing the bucket, key, lambda, and filepath flags from the command line
	filepath := flag.String("filepath", "", "Path to the file that we want to upload to AWS")
	bucket := flag.String("bucket", "", "Name of the S3 bucket that we want to upload the file to")
	key := flag.String("key", "", "Name of the key that we want to upload the file to")
	// lambdaName := flag.String("lambda", "", "Name of the Lambda function that we want to notify")
	listBuckets := flag.Bool("list-buckets", false, "List all S3 buckets")
	listFunctions := flag.Bool("list-functions", false, "List all Lambda functions")
	listKeys := flag.Bool("keys", false, "List all keys in an S3 bucket")
	flag.Parse()

	// If a bucket name and a file has been provided, upload the file to the bucket:
	if *bucket != "" && *filepath != "" {
		// Uploading the file to the s3 bucker using a goroutine
		s3Upload := make(chan error)

		go func() {
			s3Upload <- uploadFile(*filepath, *bucket, *key)
		}()

		err := <-s3Upload
		if err != nil {
			fmt.Printf("Error uploading file: %v", err)
			return
		}

		fmt.Println("File uploaded successfully")
	}

	// If the list-buckets flag is true, list the S3 buckets
	if *listBuckets {
		listBucketsChannel := make(chan error)
		go func() {
			listBucketsChannel <- listS3Buckets()
		}()
		listBucketsError := <-listBucketsChannel
		if listBucketsError != nil {
			fmt.Printf("Error listing Lambda functions: %v", listBucketsError)
			return
		}
	}

	// If the list-functions flag is true, list the Lambda functions
	if *listFunctions {
		listFunctionsChannel := make(chan error)
		go func() {
			listFunctionsChannel <- listLambdaFunctions()
		}()
		listFunctionsError := <-listFunctionsChannel
		if listFunctionsError != nil {
			fmt.Printf("Error listing Lambda functions: %v", listFunctionsError)
			return
		}
	}

	type keysResult struct {
		keys []string
		err  error
	}

	if *listKeys && *bucket != "" {
		resultCh := make(chan keysResult)
		go func() {
			ks, err := getKeys(*bucket)
			resultCh <- keysResult{keys: ks, err: err}
		}()

		result := <-resultCh
		if result.err != nil {
			fmt.Printf("Error listing keys: %v", result.err)
			return
		}
		fmt.Println(result.keys)
	}

	return
}
