package utility

import (
	"fmt"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/model"
	"os"
)

//Dependencies that are NOT required by the service, but might be used
var Dependencies = []string{}

//Cfg Config Model
var Cfg model.Config

//Exit the program incase of unacceptable errors
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
