package services

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//DownloadFile This function will copy the file from S3 bucket to destination location provided by user
//in the command line argument.It will create the director under this destination location with the
//same name as S3 bucket and then it'll copy the uploaded file from S3 to this directory
func DownloadFile(downloader *s3manager.Downloader, bucket, key *string, destLocation string) (*os.File, int64, error) {
	log.Printf("DownloadFile:: Started\r\n")
	if bucket == nil || key == nil {
		return nil, 0, nil
	}
	decodedKeyValue, err := url.QueryUnescape(*key)
	if err != nil {
		log.Fatal(err)
		return nil, 0, err
	}
	if strings.HasSuffix(decodedKeyValue, "/") {
		log.Printf("DownloadFile:: Directory create event, ignoring it: %q\r\n", decodedKeyValue)
		return nil, 0, nil
	}

	//If destination director doesn't exists with bucket name then create one
	saveDirectory := destLocation + "/" + *bucket
	if _, err := os.Stat(saveDirectory); os.IsNotExist(err) {
		os.Mkdir(saveDirectory, os.ModePerm)
		log.Printf("DownloadFile:: Created Directory on the name of bucket: %q\r\n", *bucket)
	}

	//create sub directories, if needed as per directory structure in S3 bucket
	decodedKeyValueSlice := strings.Split(decodedKeyValue, "/")
	for i := 0; i <= len(decodedKeyValueSlice)-2; i++ {
		saveDirectory = saveDirectory + "/" + decodedKeyValueSlice[i]
		if _, err := os.Stat(saveDirectory); os.IsNotExist(err) {
			os.Mkdir(saveDirectory, os.ModePerm)
			log.Printf("DownloadFile:: Created sub Directory under main directory with bucket name, with the name of: %q\r\n", decodedKeyValueSlice[i])
		}
	}

	//create directory based on unix timestamp
	saveDirectory = saveDirectory + "/" + time.Now().Format("2006-01-02_15-04-05")
	if _, err := os.Stat(saveDirectory); os.IsNotExist(err) {
		os.Mkdir(saveDirectory, os.ModePerm)
		log.Printf("DownloadFile:: Created sub Directory with unix timestmp \r\n")
	}

	// if file doesn't exist in destination directory
	saveFile := saveDirectory + "/" + decodedKeyValueSlice[len(decodedKeyValueSlice)-1]
	if _, err := os.Stat(saveFile); os.IsExist(err) {
		log.Printf("DownloadFile:: File %q already exists. Reprocessing it\r\n", decodedKeyValue)
	}

	file, err := os.Create(saveFile) //pass parameter 'location' where the file needs to be saved
	if err != nil {
		log.Printf("DownloadFile:: Failed to open or create or update file %q on local machine %v\r\n", saveFile, err)
		return nil, 0, err
	}
	defer file.Close()

	//Copy file from S3 to given file on EC2
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    aws.String(decodedKeyValue),
		})
	if err != nil {
		log.Printf("DownloadFile:: Failed to copy the file %q %v\r\n", decodedKeyValue, err)
		file.Close()
		os.Remove(saveFile)
		os.Remove(saveDirectory)
		log.Printf("DownloadFile:: Removed file %q\r\n", saveFile)
		return nil, -1, err
	}
	if numBytes == 0 {
		log.Printf("DownloadFile:: %q is file with size: %q bytes. Ignoring and removing it", file.Name(), numBytes)
		file.Close()
		os.Remove(saveFile)
		os.Remove(saveDirectory)
		log.Printf("DownloadFile:: Removed file %q\r\n", saveFile)
		return nil, -1, nil
	}
	log.Printf("DownloadFile:: %q -> File is sucessfully copied, returning\r\n", decodedKeyValue)
	log.Printf("DownloadFile:: Ended\r\n")
	return file, numBytes, nil
}
