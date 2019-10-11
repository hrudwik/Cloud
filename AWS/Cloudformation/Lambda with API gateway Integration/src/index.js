var arr = [
    "Hello, Warm welcome",
    "Now u r at ryt place",
    "Jus do it"
    ];

exports.handler = async (event) => {
    //Sample Lambda to display one of the above 3 messages
    var message = arr[(Math.floor(Math.random()*10))%3];
	
	response = {
            'isBase64Encoded': False,
            'statusCode': 200,
            'headers': {},
            'multiValueHeaders': {},
            'body': 'Hello, World!'
          }
          return response
	
    return message;
};