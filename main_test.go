package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func Reserve(router *gin.Engine, writer http.ResponseWriter, req *http.Request) {
	router.ServeHTTP(writer, req)
}

func BenchmarkTestReserve(b *testing.B) {
	router := buildRouter()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := newMockWriter()
			bodyBytes, _ := json.Marshal(gin.H{
				"external_reference": "1234",
				"mode":               "total",
				"reason":             "reserve_for_payment",
				"amount":             2500,
			})
			req, _ := http.NewRequest("POST", "/api/users/1/reserve", bytes.NewReader(bodyBytes))
			req.Header = map[string][]string{
				"X-Client-Id":       {"1234"},
				"X-Idempotency-Key": {"1234"},
			}
			Reserve(router, w, req)
		}
	})
}

type mockWriter struct {
	headers http.Header
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		http.Header{},
	}
}

func (m *mockWriter) Header() (h http.Header) {
	return m.headers
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockWriter) WriteHeader(code int) {}
