package awsPractice

import (
	"flag"
)

func main() {
	// Parsing the bucket, key, lambda, and filepath flags from the command line
	filepath := flag.String("filepath", "", "Path to the file that we want to upload to AWS")
	bucket := flag.String("bucket", "", "Name of the S3 bucket that we want to upload the file to")
	key := flag.String("key", "", "Name of the key that we want to upload the file to")
	lambdaName := flag.String("lambda", "", "Name of the Lambda function that we want to notify")
	flag.Parse()

}
