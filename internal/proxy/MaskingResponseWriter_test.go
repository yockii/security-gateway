package proxy

import (
	"fmt"
	"net/http/httptest"
	"security-gateway/pkg/server"
	"testing"
)

func TestMaskingResponseWriter_Write(t *testing.T) {
	dataStr := `{
	"data1": "1:2,3{4}5\"6[7]8 9\r\n",
	"data2": [[[120.1,30.1]]]
}`
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")

	maskingFields := []*server.DesensitizeField{
		{
			Name:                  "name",
			Level1DesensitizeRule: "each-*",
		},
		{
			Name:                  "age",
			Level1DesensitizeRule: "all-*",
		},
		{
			Name:                  "aliases",
			Level1DesensitizeRule: "start-**",
		},
		{
			Name:                  "province",
			Level1DesensitizeRule: "middle-****",
		},
		{
			Name:                  "data",
			Level1DesensitizeRule: "all-*",
		},
	}

	m := NewMaskingResponseWriter(w, maskingFields, 1)
	n, err := m.Write([]byte(dataStr))
	fmt.Println(n, err)

	fmt.Println(w.Body.String())
}
