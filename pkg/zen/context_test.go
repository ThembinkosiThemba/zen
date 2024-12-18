package zen

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req)

	if c.Writer == nil {
		t.Error("Writer should not be nil")
	}
	if c.Request != req {
		t.Error("Request not properly set")
	}
	if c.Index != -1 {
		t.Error("Index should be initialized to -1")
	}
	if len(c.Params) != 0 {
		t.Error("Params should be initialized empty")
	}
}

func TestContext_Next(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req)

	order := []string{}
	handler1 := func(c *Context) {
		order = append(order, "1")
	}
	handler2 := func(c *Context) {
		order = append(order, "2")
	}
	handler3 := func(c *Context) {
		order = append(order, "3")
	}

	c.Handlers = []HandlerFunc{handler1, handler2, handler3}
	c.Next()

	if len(order) != 3 {
		t.Error("All handlers should be executed")
	}
	if strings.Join(order, "") != "123" {
		t.Error("Handlers executed in wrong order")
	}
}

func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req)

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	testData := testStruct{
		Name: "test",
		Age:  25,
	}

	c.JSON(http.StatusOK, testData)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type header should be application/json")
	}

	var result testStruct
	err := json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if result.Name != testData.Name || result.Age != testData.Age {
		t.Error("Response data doesn't match input data")
	}
}

func TestContext_Text(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	c := NewContext(w, req)

	c.Text(http.StatusOK, "Hello %s", "World")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Error("Content-Type header should be text/plain")
	}

	expected := "Hello World"
	if got := w.Body.String(); got != expected {
		t.Errorf("Expected body %q, got %q", expected, got)
	}
}
func TestContext_BindJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr error
	}{
		{
			name:    "Valid JSON",
			body:    `{"name":"test","age":25}`,
			wantErr: nil,
		},
		{
			name:    "Invalid JSON",
			body:    `{"name":"test"`,
			wantErr: ErrBadJSON,
		},
		{
			name:    "Empty body",
			body:    "",
			wantErr: ErrEmptyBody,
		},
	}

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			c := NewContext(httptest.NewRecorder(), req)

			var result testStruct
			err := c.ParseJSON(&result)

			if err != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContext_ClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "1.1.1.1",
			},
			remoteAddr: "2.2.2.2",
			want:       "1.1.1.1",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "3.3.3.3",
			},
			remoteAddr: "2.2.2.2",
			want:       "3.3.3.3",
		},
		{
			name:       "RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "2.2.2.2",
			want:       "2.2.2.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			c := NewContext(httptest.NewRecorder(), req)
			if got := c.GetClientIP(); got != tt.want {
				t.Errorf("ClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_Context(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
	defer cancel()

	req = req.WithContext(ctx)
	c := NewContext(httptest.NewRecorder(), req)

	// Test Deadline
	_, ok := c.Deadline()
	if !ok {
		t.Error("Expected deadline to be set")
	}

	// Test Done
	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should have timed out")
	case <-c.Done():
		if c.Err() != context.DeadlineExceeded {
			t.Error("Expected DeadlineExceeded error")
		}
	}

	// Test WithValue
	key := "test-key"
	value := "test-value"
	newCtx := c.WithValue(key, value)

	if newCtx.Value(key) != value {
		t.Error("WithValue not working as expected")
	}
}
