package model

//MessageDetails this is a struct to store SQS read messages
type MessageDetails struct {
	ReceiptHandle *string
	Bucket        *string
	Key           *string
}
