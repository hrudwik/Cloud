package processor

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/services"
)

// Process is Responsible to copy the files and invoke process
func Process(queuename, destLocation, configFileName string, sess *session.Session, SQSConn *sqs.SQS,
	downloader *s3manager.Downloader, QueueURL *string) {
	log.Println("** Processor::Process Inside process() ", queuename, destLocation, " **")

	processSQSMessages(queuename, destLocation, configFileName, sess, SQSConn, downloader, QueueURL)
}

//processSQSMessages This is the core routine which will read messages, process the  and
//delete them once processed
func processSQSMessages(name, destLocation, configFileName string, sess *session.Session, SQSConn *sqs.SQS,
	downloader *s3manager.Downloader, QueueURL *string) {
	for {
		log.Printf("processSQSMessages:: Job Started\r\n")

		//Receive Message from SQS
		messageDetails, err := services.ReceiveFromQueue(SQSConn, QueueURL, name)
		if err != nil {
			log.Fatalf("processSQSMessages::Error: Failed to read message from queue %q, %v.\r\n", name, err)
			continue
		}
		//Sleep for 5 seconds, if there are no messages
		if len(messageDetails) == 0 {
			break
		}

		for i, msgDetails := range messageDetails {
			if msgDetails.Bucket != nil && msgDetails.Key != nil {
				//Download file from S3 to EC2
				file, numBytes, err := services.DownloadFile(downloader, msgDetails.Bucket, msgDetails.Key, destLocation)
				if err != nil {
					log.Printf("processSQSMessages:: Failed to download the file %v\r\n", err)
					//Need to Delete message incase of zero byte file
					if numBytes == -1 {
						log.Printf("processSQSMessages:: File with size 0 bytes is uploaded in bucket %q. Ignoring it", *msgDetails.Bucket)
						err = services.DeleteMessage(SQSConn, QueueURL, msgDetails.ReceiptHandle)
						if err != nil {
							log.Printf("processSQSMessages:: Failed to delete the message %v\r\n", err)
							continue
						} else {
							log.Printf("processSQSMessages:: Sucessfully deleted the message after execution %d\r\n", i)
						}
					}
					continue
				} else {
					if file != nil {
						log.Printf("processSQSMessages:: Downloaded file %q %d %q\r\n", file.Name(), numBytes, "bytes")

						//Calling this function asyncronously which will fork the relevant program
						if numBytes > 0 {
							go services.CallRelevantProcess(file, *msgDetails.Bucket, configFileName)
						} else {
							log.Printf("processSQSMessages:: %q File with size 0 bytes is uploaded in bucket %q. Ignoring it", file.Name(), *msgDetails.Bucket)
						}
					} else {
						if numBytes <= 0 {
							log.Printf("processSQSMessages:: File with size 0 bytes is uploaded in bucket %q. Ignoring it", *msgDetails.Bucket)
						}
					}
					//Delete the processed queue message
					err = services.DeleteMessage(SQSConn, QueueURL, msgDetails.ReceiptHandle)
					if err != nil {
						log.Printf("processSQSMessages:: Failed to delete the message %v\r\n", err)
						continue
					} else {
						log.Printf("processSQSMessages:: Sucessfully deleted the message after execution %d\r\n", i)
					}
				}
			}
			log.Printf("Job Finished\r\n")
		}
	}
}
