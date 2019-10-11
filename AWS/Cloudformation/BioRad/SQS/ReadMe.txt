nested_template s3 url:
https://templatebucketdhc.s3.us-east-2.amazonaws.com/sqs_final.yaml


CLI Command: (Execute this from root directory)
aws cloudformation create-stack --stack-name dhcsqscfnconv --template-body file://templates/template.yaml --parameters file://parameters/params.json --tags file://parameters/tags.json