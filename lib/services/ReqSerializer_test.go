package services

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestReqSerializer_encode(t *testing.T) {
	t.Run("blank", func(t *testing.T) {
		rs, err := NewReqSerializer()
		assert.Nil(t, err)
		assert.NotNil(t, rs)

		expected := `{"path":"/query","method":"GET","header":{},"query":{},"params":null}`

		req, req_err := http.NewRequest("GET", "/query", nil)
		assert.Nil(t, req_err)
		assert.NotNil(t, req)

		enc, enc_err := rs.Encode(req, true)
		assert.Nil(t, enc_err)
		assert.NotNil(t, enc)

		fmt.Sprintf("JSON: %s\n", string(enc))
		assert.JSONEq(t, expected, string(enc))
	})
	t.Run("nonempty", func(t *testing.T) {
		rs, err := NewReqSerializer()
		assert.Nil(t, err)
		assert.NotNil(t, rs)

		expected := `{"path":"/query","method":"GET","header":{"X-Tags":["feature1","feature2"]},"query":{"id":["1001"],"status":["closed"],"type":["vehicle","car"]},"params":null}`

		req, req_err := http.NewRequest("GET", "/query?id=1001&type=vehicle&type=car&status=closed", nil)
		assert.Nil(t, req_err)
		assert.NotNil(t, req)

		req.Header.Add("x-tags", "feature1")
		req.Header.Add("X-Tags", "feature2")

		enc, enc_err := rs.Encode(req, true)
		assert.Nil(t, enc_err)
		assert.NotNil(t, enc)

		fmt.Sprintf("JSON: %s\n", string(enc))
		assert.JSONEq(t, expected, string(enc))
	})
}

func TestReqSerializer_decode(t *testing.T) {
	t.Run("blank", func(t *testing.T) {
		rs, err := NewReqSerializer()
		assert.Nil(t, err)
		assert.NotNil(t, rs)

		expected := &RequestPacket{
			Header: http.Header{},
			Query: url.Values{},
		}

		dec, dec_err := rs.Decode([]byte(`{"header":{},"query":{},"params":null}`))
		assert.Nil(t, dec_err)
		assert.NotNil(t, dec)

		fmt.Sprintf("JSON: %v\n", dec)
		assert.EqualValues(t, expected, dec)
	})
	t.Run("nonempty", func(t *testing.T) {
		rs, err := NewReqSerializer()
		assert.Nil(t, err)
		assert.NotNil(t, rs)

		expected := &RequestPacket{
			Header: http.Header{
				"X-Tags": []string {"feature1","feature2"},
			},
			Query: url.Values{
				"id": []string {"1001"},
				"status": []string {"closed"},
				"type": []string {"vehicle","car"},
			},
			Params: nil,
		}

		dec, dec_err := rs.Decode([]byte(`{"header":{"X-Tags":["feature1","feature2"]},"query":{"id":["1001"],"status":["closed"],"type":["vehicle","car"]},"params":null}`))
		assert.Nil(t, dec_err)
		assert.NotNil(t, dec)

		fmt.Sprintf("JSON: %v\n", dec)
		assert.EqualValues(t, expected, dec)
	})
}
