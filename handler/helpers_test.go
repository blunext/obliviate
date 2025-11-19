package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetStatusAndHeader(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		prodEnv  bool
		wantType string
	}{
		{
			name:     "200 OK in production",
			status:   http.StatusOK,
			prodEnv:  true,
			wantType: "application/json",
		},
		{
			name:     "404 Not Found in development",
			status:   http.StatusNotFound,
			prodEnv:  false,
			wantType: "application/json",
		},
		{
			name:     "500 Internal Server Error",
			status:   http.StatusInternalServerError,
			prodEnv:  true,
			wantType: "application/json",
		},
		{
			name:     "400 Bad Request",
			status:   http.StatusBadRequest,
			prodEnv:  false,
			wantType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			setStatusAndHeader(w, tt.status, tt.prodEnv)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.status, result.StatusCode, "status code should match")
			assert.Equal(t, tt.wantType, result.Header.Get("Content-Type"), "content-type should be application/json")
		})
	}
}

func TestJsonFromStruct(t *testing.T) {
	tests := []struct {
		name       string
		input      interface{}
		wantErr    bool
		validateFn func(*testing.T, []byte)
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 42,
			},
			wantErr: false,
			validateFn: func(t *testing.T, result []byte) {
				var output map[string]interface{}
				err := json.Unmarshal(result, &output)
				require.NoError(t, err)
				assert.Equal(t, "test", output["name"])
				assert.Equal(t, float64(42), output["value"])
			},
		},
		{
			name: "struct with byte array",
			input: struct {
				Data []byte `json:"data"`
			}{
				Data: []byte("hello world"),
			},
			wantErr: false,
			validateFn: func(t *testing.T, result []byte) {
				var output map[string]interface{}
				err := json.Unmarshal(result, &output)
				require.NoError(t, err)
				assert.NotNil(t, output["data"])
			},
		},
		{
			name: "empty struct",
			input: struct {
			}{},
			wantErr: false,
			validateFn: func(t *testing.T, result []byte) {
				assert.Equal(t, "{}", string(result))
			},
		},
		{
			name: "nested struct",
			input: struct {
				Outer string `json:"outer"`
				Inner struct {
					Field string `json:"field"`
				} `json:"inner"`
			}{
				Outer: "outer-value",
				Inner: struct {
					Field string `json:"field"`
				}{
					Field: "inner-value",
				},
			},
			wantErr: false,
			validateFn: func(t *testing.T, result []byte) {
				var output map[string]interface{}
				err := json.Unmarshal(result, &output)
				require.NoError(t, err)
				assert.Equal(t, "outer-value", output["outer"])
				inner, ok := output["inner"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "inner-value", inner["field"])
			},
		},
		{
			name: "struct with omitempty",
			input: struct {
				Present string `json:"present"`
				Empty   string `json:"empty,omitempty"`
			}{
				Present: "value",
				Empty:   "",
			},
			wantErr: false,
			validateFn: func(t *testing.T, result []byte) {
				var output map[string]interface{}
				err := json.Unmarshal(result, &output)
				require.NoError(t, err)
				assert.Equal(t, "value", output["present"])
				_, exists := output["empty"]
				assert.False(t, exists, "empty field should be omitted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonFromStruct(context.Background(), tt.input)

			assert.NotNil(t, result)

			if tt.validateFn != nil {
				tt.validateFn(t, result)
			}
		})
	}
}

func TestJsonFromStruct_InvalidType(t *testing.T) {
	// Test with channel (which cannot be marshaled to JSON)
	invalidStruct := struct {
		Ch chan int
	}{
		Ch: make(chan int),
	}

	// Should not panic, but log error and return nil
	// The function logs the error but returns the nil result from json.Marshal
	assert.NotPanics(t, func() {
		jsonFromStruct(context.Background(), invalidStruct)
	}, "should not panic on marshal error")
}

func TestFinishRequestWithErr(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		status  int
		prodEnv bool
	}{
		{
			name:    "bad request in production",
			msg:     "invalid input",
			status:  http.StatusBadRequest,
			prodEnv: true,
		},
		{
			name:    "internal error in development",
			msg:     "database connection failed",
			status:  http.StatusInternalServerError,
			prodEnv: false,
		},
		{
			name:    "not found error",
			msg:     "resource not found",
			status:  http.StatusNotFound,
			prodEnv: true,
		},
		{
			name:    "unauthorized error",
			msg:     "authentication required",
			status:  http.StatusUnauthorized,
			prodEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			finishRequestWithErr(context.Background(), w, tt.msg, tt.status, tt.prodEnv)

			result := w.Result()
			defer result.Body.Close()

			// Verify status code
			assert.Equal(t, tt.status, result.StatusCode)

			// Verify Content-Type header
			assert.Equal(t, "application/json", result.Header.Get("Content-Type"))

			// Verify body is empty
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Empty(t, body)
		})
	}
}

func TestFinishRequestWithWarn(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		status  int
		prodEnv bool
	}{
		{
			name:    "not found warning in production",
			msg:     "Message not found",
			status:  http.StatusNotFound,
			prodEnv: true,
		},
		{
			name:    "deprecation warning in development",
			msg:     "API endpoint deprecated",
			status:  http.StatusOK,
			prodEnv: false,
		},
		{
			name:    "validation warning",
			msg:     "Some fields were ignored",
			status:  http.StatusOK,
			prodEnv: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			finishRequestWithWarn(context.Background(), w, tt.msg, tt.status, tt.prodEnv)

			result := w.Result()
			defer result.Body.Close()

			// Verify status code
			assert.Equal(t, tt.status, result.StatusCode)

			// Verify Content-Type header
			assert.Equal(t, "application/json", result.Header.Get("Content-Type"))

			// Verify body is empty
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Empty(t, body)
		})
	}
}

func TestHelpers_Integration(t *testing.T) {
	// Test that helpers work together correctly in a realistic scenario

	t.Run("successful response flow", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Simulate successful response
		setStatusAndHeader(w, http.StatusOK, true)

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{
			Message: "success",
			Code:    200,
		}

		jsonBytes := jsonFromStruct(context.Background(), response)
		_, err := w.Write(jsonBytes)
		require.NoError(t, err)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Equal(t, "application/json", result.Header.Get("Content-Type"))

		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		var decoded map[string]interface{}
		err = json.Unmarshal(body, &decoded)
		require.NoError(t, err)
		assert.Equal(t, "success", decoded["message"])
		assert.Equal(t, float64(200), decoded["code"])
	})

	t.Run("error response flow", func(t *testing.T) {
		w := httptest.NewRecorder()

		finishRequestWithErr(context.Background(), w, "test error", http.StatusBadRequest, false)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
		assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
	})
}

func TestHelpers_ConcurrentSafety(t *testing.T) {
	// Verify helpers are safe to use concurrently
	t.Run("concurrent jsonFromStruct calls", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(val int) {
				defer func() { done <- true }()

				data := struct {
					Value int `json:"value"`
				}{
					Value: val,
				}

				result := jsonFromStruct(context.Background(), data)
				assert.NotNil(t, result)
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
