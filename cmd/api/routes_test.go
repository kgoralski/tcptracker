package api

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndpoints(t *testing.T) {
	router := NewRouter(chi.NewRouter(), prometheus.NewRegistry())
	router.Routes()
	w := httptest.NewRecorder()

	tests := []struct {
		name       string
		statusCode int
		want       interface{}
	}{
		{
			name:       "/health",
			statusCode: http.StatusOK,
			want: &Response{
				Message:    "healthy",
				StatusText: http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
			},
		},
		{
			name:       "/metrics",
			statusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest(http.MethodGet, tt.name, nil)
			if err != nil {
				t.Fatal(err)
			}
			router.mux.ServeHTTP(w, r)
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Less(t, 0, w.Body.Len())
			if tt.want != nil {
				var actualResponse Response
				err = json.NewDecoder(w.Body).Decode(&actualResponse)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.want, &actualResponse)
			}
		})
	}
}
