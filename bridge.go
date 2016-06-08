package kafkaduct

import (
	"net/http"
	"encoding/json"
)

type Bridge struct {
	client *http.Client
	timer chan struct{}
	topics []string
	reqBuilder func(string) http.Request
	messages chan HttpBridgeResponse{}
}

type HttpBridgeResponse struct {
	messages []HttpBridgedMessage
}

type HttpBridgedMessage struct {
	Body string `json:"body"`
	MessageType string `json:"messageType"`
	// TODO rest of fields
}

func NewBridge(appConfig *AppConfig) Bridge {
	

	bridge := Bridge{} 

	bridge.client := &http.Client{
		CheckRedirect: redirectPolicyFunc
	}

	bridge.timer := make(chan struct{})

}

func (b *Bridge) Run() {
	for _,topic range b.topics {
		go b.loop(topic)
	}
}

func (b *Bridge) loop(topic string) {
	for {
		select case  <- b.timer:
			messages, err := b.doPoll(topic)
			b.forward(messages)
	}
}

func doPoll(topic) []HttpBridgedMessage, error {
	req := b.reqBuilder(topic)
	resp, err := client.Do(req)
	if err != nil {
		log.Log("ERROR","HTTP Request Failed", "error", err, "url", req.URL)
		return nil, "HTTP Request Failed"
	}

	if resp.StatusCode == 204 {
		return doPoll(topic), nil

	} else if resp.StatusCode == 200 {

		msgs := HttpBridgeResponse{}

		if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
			log.Log("ERROR","Decoding JSON Response", "error", err, "url", req.URL)
			return
		}

		return HttpBridgedMessage.messages
	}

	return nil, "Non-Success status"
}

func NewRequestBuilder(appConfig *AppConfig) func(string) {
	return func(topic) {
		u := url.ParseUrl(appConfig.Bridge.Url)

		// TODO need to replace place-holders in URL, i.e. {topic}, {groupId}, {clientId}
		req, err := http.NewRequest("GET", u)

		if err != nil {
			panic(err)
		}

		req.Header.Add("X-Api-Key", appConfig.Bridge.ApiKey)
		req.Header.Add("Accept", "application/json")

		return req
	}
}
