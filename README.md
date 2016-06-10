# kafkaduct

Sends [`kafka-httpbridge`](http://git.svc.ft.com:8080/projects/MXT/repos/kafka-http-bridge/browse)-style `POST` messages to a kafka log.

## Checkout

Clone under your `GOPATH` as `$GOPATH/src/github.com/Financial-Times/kafkaduct`:

    $ git clone git@github.com:Financial-Times/kafkaduct.git

## Fetch dependencies

If you don't already have it installed, install [`godep`](https://github.com/tools/godep):

    $ go get github.com/tools/godep

Then, from `kafkaduct` repo directory:

    $ godep get

## Building and Running locally

> You will also need kafka running locally, e.g. using the `do env` script from  [`om-apps-environments`](http://git.svc.ft.com:8080/projects/OM/repos/om-apps-environments/browse).


    $ . sourceme     # sets dummy env-vars
    $ cd kafkaduct
    $ go build && ./kafkaduct

Do a test request:
```
curl -X POST -H "Content-Type: application/json" -H "X-Api-Key: someKey" -d '{
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
' "https://localhost:8080/write"

```

## Before commit

1. `godep save`
2. `godep save ./...`

This will refresh the `Godeps.json` file, removing unused dependencies.
