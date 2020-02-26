package model

//SqsBody format of SQS message body as a struct
type SqsBody struct {
	Records []struct {
		EventName string
		S3        struct {
			Bucket struct {
				Name string
			}
			Object struct {
				Key  string
				Size int
			}
		}
	}
}