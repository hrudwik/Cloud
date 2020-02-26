package services

import (
	"log"
	"os"
)

//ValidateQueueNameAndDestinationPath This fuction will create the user session based on either the EC2 role or User redentials
//provided to the EC2 instance as environment variables or credential file (~/.AWS/credentials
func ValidateQueueNameAndDestinationPath(queuename, destLocation string) {
	if len(queuename) == 0 {
		log.Fatalln("ValidateQueueNameAndDestinationPath:: Queue name required \r\n")
	}

	if len(destLocation) == 0 {
		log.Fatalln("ValidateQueueNameAndDestinationPath:: Destination directory location is required \r\n")
	}

	//Validate provided destination path
	if _, err := os.Stat(destLocation); os.IsNotExist(err) {
		log.Fatalln("ValidateQueueNameAndDestinationPath:: Destination directory path provided is invalid \r\n")
	}
}
