package services

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/model"
)

//ReceiveFromQueue This function will receive messages from SQS, it uses long polling mechanishm.
//It can read atmost 10 messages per poll.
func ReceiveFromQueue(SQSConn *sqs.SQS, QueueURL *string, name string) ([]model.MessageDetails, error) {
	log.Printf("ReceiveFromQueue:: Started\r\n")
	// Receive message params
	receiveParams := &sqs.ReceiveMessageInput{
		QueueUrl: QueueURL,
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(10),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds:   aws.Int64(20), //Timeout for fetch message from queue
		VisibilityTimeout: aws.Int64(30), //Time till the received queue will not be visible
	}

	// Receive a message from the SQS queue with long polling enabled.
	result, err := SQSConn.ReceiveMessage(receiveParams)
	if err != nil {
		log.Printf("ReceiveFromQueue:: Unable to receive message from queue %q\r\n", name)
		return nil, err
	}

	var n int = len(result.Messages)
	log.Printf("ReceiveFromQueue:: Received %d messages.\r\n", n)
	messageDetails := make([]model.MessageDetails, n)

	for i := 0; i < n; i++ {
		receiptHandle := result.Messages[i].ReceiptHandle
		body := result.Messages[i].Body
		// unmarshal sqs message body
		data := &model.SqsBody{}
		err = json.Unmarshal([]byte(*body), &data)
		if err != nil {
			log.Printf("ReceiveFromQueue:: Failed to parse message data\r\n")
		}
		if len(data.Records) == 0 {
			log.Printf("ReceiveFromQueue:: Irrelavent SQS message, ignoring message\r\n")
			continue
		}
		bucket := &data.Records[0].S3.Bucket.Name
		key := &data.Records[0].S3.Object.Key
		msgDetails := model.MessageDetails{
			ReceiptHandle: receiptHandle,
			Bucket:        bucket,
			Key:           key,
		}
		messageDetails[i] = msgDetails
	}

	log.Printf("ReceiveFromQueue:: Ended\r\n")
	return messageDetails, nil
}

//DeleteMessage This function will delete the message from SQS
func DeleteMessage(SQSConn *sqs.SQS, QueueURL, receiptHandle *string) error {
	params := &sqs.DeleteMessageInput{
		QueueUrl:      QueueURL,
		ReceiptHandle: receiptHandle,
	}
	_, err := SQSConn.DeleteMessage(params)
	if err != nil {
		return err
	}
	return nil
}
