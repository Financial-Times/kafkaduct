# kafkaduct

1. `go build && go install .`
2. `export DYNO="1"`
3. `export KAFKA_URL="http://localhost:8080/kafka"`
4. `export PORT="8080"`
5. `export API_KEY="<someKey>"`
6. `./kafkaduct`

Do a test request
```
curl -X POST -H "Content-Type: application/json" -H "X-Api-Key: <API_KEY>" -d '{
    "foo": "bar"
}
' "http://localhost:8080/events"

```
