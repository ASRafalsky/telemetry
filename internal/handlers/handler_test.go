package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"

	testrepo "github.com/ASRafalsky/telemetry/internal/repository"
	"github.com/ASRafalsky/telemetry/internal/transport"
	"github.com/ASRafalsky/telemetry/pkg/cache"
)

type testData struct {
	name          string
	data          transport.Metrics
	expStatusCode int
	expResponse   transport.Metrics
}

type testRequest struct {
	req           *http.Request
	resp          *httptest.ResponseRecorder
	expStatusCode int
}

const url = "/update/"

func TestJSONPostHandler(t *testing.T) {
	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	handler := http.HandlerFunc(JSONPostHandler(repo, SetDataTo))

	for _, tc := range getTestDataList() {
		t.Run(tc.name, func(t *testing.T) {
			buf, err := easyjson.Marshal(tc.data)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
			require.NoError(t, err)

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			require.Equal(t, tc.expStatusCode, resp.Code)
			if tc.expStatusCode == http.StatusOK {
				require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
				body := resp.Body.Bytes()
				require.NotEmpty(t, body)
				m := transport.Metrics{}
				require.NoError(t, easyjson.Unmarshal(body, &m))
				require.Equal(t, tc.expResponse, m)
			}
		})
	}
}

func BenchmarkJSONPostHandler(b *testing.B) {
	dataList := getTestDataList()

	tt := make([]testRequest, len(dataList))

	for i, tc := range getTestDataList() {
		buf, err := easyjson.Marshal(tc.data)
		require.NoError(b, err)

		tt[i].req, err = http.NewRequest("POST", url, bytes.NewReader(buf))
		require.NoError(b, err)
		tt[i].expStatusCode = tc.expStatusCode
		tt[i].resp = httptest.NewRecorder()
	}

	repo := testrepo.NewExtendedRepository(cache.New[string, []byte]())
	handler := http.HandlerFunc(JSONPostHandler(repo, SetDataTo))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n := range tt {
			handler.ServeHTTP(tt[n].resp, tt[n].req)
			require.Equal(b, tt[n].expStatusCode, tt[n].resp.Code)
		}
	}
}

func getTestDataList() []testData {
	gaugeVal := 123.5
	counterVal := int64(123)
	updatedCounterVal := counterVal + counterVal

	return []testData{
		{
			name: "correct_gauge_update_1",
			data: transport.Metrics{
				MType: "gauge",
				ID:    "g_value0",
				Value: &gaugeVal,
			},
			expStatusCode: http.StatusOK,
			expResponse: transport.Metrics{
				MType: "gauge",
				ID:    "g_value0",
				Value: &gaugeVal,
			},
		},
		{
			name: "correct_gauge_update_2",
			data: transport.Metrics{
				MType: "gauge",
				ID:    "g_value1",
				Value: &gaugeVal,
			},
			expStatusCode: http.StatusOK,
			expResponse: transport.Metrics{
				MType: "gauge",
				ID:    "g_value1",
				Value: &gaugeVal,
			},
		},
		{
			name: "correct_counter_update_1",
			data: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
				Delta: &counterVal,
			},
			expStatusCode: http.StatusOK,
			expResponse: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
				Delta: &counterVal,
			},
		},
		{
			name: "correct_counter_update_2",
			data: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
				Delta: &counterVal,
			},
			expStatusCode: http.StatusOK,
			expResponse: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
				Delta: &updatedCounterVal,
			},
		},
		{
			name: "incorrect_counter_update_1",
			data: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
				Value: &gaugeVal,
			},
			expStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect_counter_update_2",
			data: transport.Metrics{
				MType: "counter",
				ID:    "c_value0",
			},
			expStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect_gauge_update_1",
			data: transport.Metrics{
				MType: "gauge",
				ID:    "g_value0",
				Delta: &counterVal,
			},
			expStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect_gauge_update_2",
			data: transport.Metrics{
				MType: "gauge",
				ID:    "g_value0",
			},
			expStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect_type",
			data: transport.Metrics{
				MType: "lol",
				ID:    "g_value0",
				Value: &gaugeVal,
			},
			expStatusCode: http.StatusBadRequest,
		},
	}
}
