package services

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type ReqSerializer struct {}

type RequestPacket struct {
	Header http.Header `json:"header"`
	Query  url.Values `json:"query"`
	Params map[string][]string `json:"params"`
}

func NewReqSerializer() (*ReqSerializer, error) {
	return &ReqSerializer{}, nil
}

func (s *ReqSerializer) Encode(r *http.Request) ([]byte, error) {
	packet := &RequestPacket{
		Header: r.Header,
		Query: r.URL.Query(),
	}
	return json.Marshal(packet)
}

func (s *ReqSerializer) Decode(data []byte) (*RequestPacket, error) {
	packet := &RequestPacket{}
	err := json.Unmarshal(data, packet)
	return packet, err
}
