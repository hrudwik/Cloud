package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//CallRelevantProcess This function will call the program to be executed based on config.json file
//It reads the program path and arguments from the json file and will execute that particular program
func CallRelevantProcess(lastProcessedFile *os.File, directoryName, configFileName string) (bytes.Buffer, error) {
	log.Printf("CallRelevantProcess::Started %q \r\n", configFileName)
	var outb, errb bytes.Buffer

	file, err := os.Open(configFileName)
	if err != nil {
		log.Printf("CallRelevantProcess::Error: Failed open config.json file %v.\r\n", err)
		return errb, err
	} else {
		defer file.Close()
		decoder := json.NewDecoder(file)
		var result map[string]interface{}
		err := decoder.Decode(&result)
		if err != nil {
			log.Printf("CallRelevantProcess::Error: Failed to read %q file %v.\r\n", configFileName, err)
			return errb, err
		}
		filePath := lastProcessedFile.Name()
		startIndex := strings.Index(filePath, directoryName)
		endIndex := strings.LastIndex(filePath, "/")
		relativeDirectoryName := filePath[startIndex:endIndex]

		//ignore unix timestamp directory
		endIndex = strings.LastIndex(relativeDirectoryName, "/")
		relativeDirectoryName = relativeDirectoryName[0:endIndex]

		log.Printf("CallRelevantProcess::RelativeDirectoryName to be searched from %q is: %q.\r\n", configFileName, relativeDirectoryName)
		if _, ok := result[relativeDirectoryName]; ok {
			progJSON := result[relativeDirectoryName].(map[string]interface{})
			program := ""
			arguments := []interface{}{}
			if _, ok := progJSON["program"]; ok {
				program = progJSON["program"].(string)
			} else {
				log.Printf("CallRelevantProcess:: program is missing in %q in value with key %q\r\n", configFileName, relativeDirectoryName)
				return errb, err
			}

			if _, ok := progJSON["arguments"]; ok {
				arguments = progJSON["arguments"].([]interface{})
			} else {
				log.Printf("CallRelevantProcess:: arguments are missing in %q in value with key %q\r\n", configFileName, relativeDirectoryName)
			}
			args := make([]string, len(arguments)+1)
			log.Printf("CallRelevantProcess:: Last processed name as first arg %q\r\n", lastProcessedFile.Name())
			args[0] = lastProcessedFile.Name()
			for i, v := range arguments {
				args[i+1] = fmt.Sprint(v)
			}
			log.Printf("CallRelevantProcess:: Last processed name as first arg %q\r\n", args[0])

			//Invoke the program with the filePath as first argument
			cmd := exec.Command(program, args...)
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			log.Printf("CallRelevantProcess::About to execute program\r\n")
			err1 := cmd.Run()
			if err1 != nil {
				log.Printf("CallRelevantProcess:: Failed to execute the program %v\r\n", err1)
				return errb, err1
			} else {
				log.Printf("CallRelevantProcess:: Finished executing the program\r\n")
			}
			log.Printf("CallRelevantProcess::out: %q err: %q\r\n", outb.String(), errb.String())
		} else {
			log.Printf("CallRelevantProcess:: could not find key %q in %q\r\n", relativeDirectoryName, configFileName)
		}
		return errb, nil
	}
}
