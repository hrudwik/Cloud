package unittest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"

	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/model"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/services"
)

var queueURL = "https://queue.amazonaws.com/80398EXAMPLE/bioradtestQueue"
var bucket = "bioradtestbucket"
var key = "test.txt"

var content = `This is the text in text file`

var destLoc = "/home/ubuntu/saveLocation"
var configFileName = "/home/ubuntu/go/S3Upload_InvokeProcessUtility/config/InvokeProcess_test-config.json"

func TestS3toEC2FileCopyUtility(t *testing.T) {
	// fake s3
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()

	// configure a dummy S3 client
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Endpoint:         aws.String(ts.URL),
		Region:           aws.String("eu-central-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	cparams := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	// Create a new dummy bucket using the CreateBucket call.
	_, err := s3Client.CreateBucket(cparams)
	if err != nil {
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Upload a new object "testobject" with the some content  to our bucket.
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(content),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	//Receive message for dummy queue
	messageDetails, err := ReceiveFromQueue(bucket, key, queueURL)
	if err != nil {
		log.Fatalf("processSQSMessages::Error: Failed to read message from queue %v.\r\n", err)
	}

	fmt.Println("TestS3toEC2FileCopyUtility:: Bucket: ", *messageDetails.Bucket)
	fmt.Println("TestS3toEC2FileCopyUtility:: Key: ", *messageDetails.Key)

	//initialize S3Downloader
	downloader := s3manager.NewDownloader(newSession)

	//Testing DownloadFile() from services module
	file, numBytes, err := services.DownloadFile(downloader, messageDetails.Bucket, messageDetails.Key, destLoc)
	if err != nil {
		fmt.Println("processSQSMessages:: Failed to download the file ", err)
		if numBytes == -1 {
			fmt.Println("processSQSMessages:: file with size 0 bytes. Ignoring it")
		} else {
			t.FailNow()
		}
	} else {
		if file != nil {
			if numBytes <= 0 {
				fmt.Println("processSQSMessages::", file.Name(), " is file with size 0 bytes. Ignoring it")
			} else {
				fmt.Println("processSQSMessages:: Downloaded file ", file.Name(), numBytes, "bytes")
				//Testing CallRelevantProcess() from services module
				errb, err := services.CallRelevantProcess(file, *messageDetails.Bucket, configFileName)
				if err != nil {
					fmt.Println("processSQSMessages:: CallRelevantProcess returned error ", err)
					t.FailNow()
				}
				if errb.String() != "" {
					fmt.Println("processSQSMessages:: Invoked process returned failure ", errb.String())
					t.FailNow()
				}
			}
		}
	}
}

//messageBody sqs message
//var messageBody = "{\"Records\":[{\"eventVersion\":\"2.1\",\"eventSource\":\"aws:s3\",\"awsRegion\":\"us-east-2\",\"eventTime\":\"2019-11-02T10:37:55.009Z\",\"eventName\":\"ObjectCreated:Put\",\"userIdentity\":{\"principalId\":\"APMVFF5NG2CUD\"},\"requestParameters\":{\"sourceIPAddress\":\"103.6.33.5\"},\"responseElements\":{\"x-amz-request-id\":\"6E55074F89EF6E73\",\"x-amz-id-2\":\"2MnNOP6KFHpiWNeQ6RezDa8BBmTlt5weXWNh5BjlxfxZieQu1OselkovYV32bOkwZdMoQJkaQj0=\"},\"s3\":{\"s3SchemaVersion\":\"1.0\",\"configurationId\":\"OnUploadSQS\",\"bucket\":{\"name\":\"bioradtestbucket\",\"ownerIdentity\":{\"principalId\":\"APMVFF5NG2CUD\"},\"arn\":\"arn:aws:s3:::bioradtestbucket\"},\"object\":{\"key\":\"Jakesh_Questions.txt\",\"size\":262,\"eTag\":\"c6cfde0062c531dc078dc60d92810354\",\"sequencer\":\"005DBD5C82F3975B70\"}}}]}"

//Random String code
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//StringWithCharset generates random string of given length
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//getNewHandle Creates new handle for SQS Message
func getNewHandle(length int) string {
	return StringWithCharset(length, charset)
}

//SQS Mock
type mockSQS struct {
	sqsiface.SQSAPI
	messages map[string][]*sqs.Message
}

//SqsBody format of SQS message body as a struct
type sQSMessages struct {
	Body          model.SqsBody
	ReceiptHandle *string
}

type Bucket struct {
	Name string
}

type Object struct {
	Key  string
	Size int
}

type S3 struct {
	Bucket Bucket
	Object Object
}

type myRecords struct {
	EventName string
	S3        S3
}

type SqsBody struct {
	Records []myRecords
}

//SendMessage will send message to our dummy SQS
func (m *mockSQS) SendMessage(in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	ReceiptHandlel := getNewHandle(10)
	m.messages[*in.QueueUrl] = append(m.messages[*in.QueueUrl], &sqs.Message{
		Body:          in.MessageBody,
		ReceiptHandle: &ReceiptHandlel,
	})
	return &sqs.SendMessageOutput{}, nil
}

//ReceiveMessage will receive message from our dummy SQS
func (m *mockSQS) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if len(m.messages[*in.QueueUrl]) == 0 {
		return &sqs.ReceiveMessageOutput{}, nil
	}
	response := m.messages[*in.QueueUrl][0:1]
	m.messages[*in.QueueUrl] = m.messages[*in.QueueUrl][1:]
	return &sqs.ReceiveMessageOutput{
		Messages: response,
	}, nil
}

//getMockSQSClient This will just mock the sqs client
func getMockSQSClient() sqsiface.SQSAPI {
	return &mockSQS{
		messages: map[string][]*sqs.Message{},
	}
}

//ReceiveFromQueue This will handle the processing SQSMessages
func ReceiveFromQueue(bucketName, fileName, queueURL string) (model.MessageDetails, error) {
	q := getMockSQSClient()
	sqsBody := &SqsBody{}
	messageBodyStruct := myRecords{
		EventName: "Event1",
		S3: S3{
			Bucket: Bucket{
				Name: bucketName,
			},
			Object: Object{
				Key:  fileName,
				Size: 262,
			},
		},
	}
	sqsBody.Records = append(sqsBody.Records, messageBodyStruct)

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(sqsBody)
	messageBody := string(reqBodyBytes.Bytes())

	QueueURL := queueURL
	q.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    &QueueURL,
	})
	message, _ := q.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: &QueueURL,
	})

	receiptHandle := message.Messages[0].ReceiptHandle
	body := message.Messages[0].Body
	log.Printf("ReceiveFromQueue:: receiptHandle %q \r\n", *receiptHandle)
	log.Printf("ReceiveFromQueue:: receiptHandle %q \r\n", *body)
	// unmarshal sqs message body
	data := &model.SqsBody{}
	err := json.Unmarshal([]byte(*body), &data)
	if err != nil {
		log.Printf("ReceiveFromQueue:: Failed to parse message data\r\n")
		return model.MessageDetails{}, err
	}
	if len(data.Records) == 0 {
		log.Printf("ReceiveFromQueue:: Irrelavent SQS message, ignoring message\r\n")
	} else {
		bucket := &data.Records[0].S3.Bucket.Name
		key := &data.Records[0].S3.Object.Key
		msgDetails := model.MessageDetails{
			ReceiptHandle: receiptHandle,
			Bucket:        bucket,
			Key:           key,
		}
		fmt.Println("Bucket: ", *msgDetails.Bucket)
		fmt.Println("Key: ", *msgDetails.Key)

		return msgDetails, nil
	}

	return model.MessageDetails{}, nil
}
