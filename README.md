# kafkaduct

1. `go build && go install .`
2. `export DYNO="1"`
3. `export KAFKA_URL="http://localhost:8080/kafka"`
4. `export PORT=":8080"`
5. `export API_KEY="<someKey>"`
6. `export KAFKA_HTTPBRIDGE_URL="https://kafka-httpbridge-gw-eu-west-1-test.memb.ft.com"`
7. `export KAFKA_HTTPBRIDGE_API_KEY="<the key>"`
8. `./kafkaduct`

Do a test request
```
curl -X POST -H "Content-Type: application/json" -d '{
    "foo": "bar"
}
' "http://localhost:8080/events"

```
