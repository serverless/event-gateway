package functions

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

// Call tries to send a payload to a target function
func (p *HTTPProperties) Call(payload []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Post(p.URL, "application/json; charset=utf-8", bytes.NewReader(payload))
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
