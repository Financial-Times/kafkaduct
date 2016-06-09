# kafkaduct

TODO: TLS setup

1. `export DYNO="1"`
2. `export KAFKA_URL="<heroku_kafka_url>"`
3. `export PORT="8080"`
4. `export API_KEY="<someKey>"`
5. `go build && go install . && ./kafkaduct`

Do a test request
```
curl -X POST -H "Content-Type: application/json" -H "X-Api-Key: <API_KEY>" -d '{
    "messages": [
        {
            "body": "Pong",
            "contentType": "text/plain",
            "messageId": "0fb133c4-68f1-47f4-9f32-985f308bd4f6",
            "messageTimestamp": "2016-06-08T15:24:57.639+0000",
            "messageType": "Ping.TestMessage",
            "originHost": "dev",
            "originHostLocation": "dev",
            "originSystemId": "kafka-http-bridge"
        }
    ],
    "topic": "membership_users_v1"
}
' "https://<host:port>/write"

```

## Before commit

1. `godep save`
2. `godep save ./...`

This will refresh the `Godeps.json` file, removing unused dependencies.
