package services

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
)

//CreateSession This fuction will create the user session based on either the EC2 role or User redentials
//provided to the EC2 instance as environment variables or credential file (~/.AWS/credentials
func CreateSession(name string) (*session.Session, *sqs.SQS, *s3manager.Downloader, *string) {
	log.Printf("CreateSession:: Started\r\n")
	metaSession, err := session.NewSession()
	if err != nil {
		log.Fatalln("processSQSMessages:: Error: Unable to establish session %q, %v.\n", name, err)
	}
	metaClient := ec2metadata.New(metaSession)
	//Fetch region
	region, err := metaClient.Region()
	if err != nil {
		log.Fatalln("processSQSMessages:: Error: Unable to fetch region %q, %v.\n", name, err)
	}

	conf := aws.NewConfig().WithRegion(region)
	sess, err := session.NewSession(conf)
	if err != nil {
		log.Fatalln("processSQSMessages:: Error: Unable to establish session with region %q, %v.\n", name, err)
	}

	// Create a SQS serviceclient.
	SQSCon := sqs.New(sess)

	//S3 downloader
	downloader := s3manager.NewDownloader(sess)

	// Need to convert the queue name into a QueueURL through GtQueueUrl() API call
	resultURL, err := SQSCon.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		log.Fatalln("processSQSMessages::Error: Unable to find queue %q, %v.\n", name, err)
	}
	QueueURL := resultURL.QueueUrl

	log.Printf("CreateSession:: Ended\r\n")
	return sess, SQSCon, downloader, QueueURL
}
