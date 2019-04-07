package services

import (
	"encoding/json"
	"net/http"
	"net/url"
	"github.com/gorilla/mux"
)

type ReqSerializer struct {}

type RequestPacket struct {
	Method *string `json:"method"`
	Path *string `json:"path"`
	Header http.Header `json:"header"`
	Query  url.Values `json:"query"`
	Params map[string]string `json:"params"`
}

func NewReqSerializer() (*ReqSerializer, error) {
	return &ReqSerializer{}, nil
}

func (s *ReqSerializer) Encode(r *http.Request, defaultUrl bool) ([]byte, error) {
	packet := &RequestPacket{
		Method: &r.Method,
		Path: &r.URL.Path,
		Header: r.Header,
		Query: r.URL.Query(),
	}
	if !defaultUrl {
		packet.Params = mux.Vars(r)
	}
	return json.Marshal(packet)
}

func (s *ReqSerializer) Decode(data []byte) (*RequestPacket, error) {
	packet := &RequestPacket{}
	err := json.Unmarshal(data, packet)
	return packet, err
}
