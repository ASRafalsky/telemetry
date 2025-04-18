package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerStatuses(t *testing.T) {
	srv := httptest.NewServer(newRouter())
	defer srv.Close()

	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}
	expHeader := http.Header{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	}
	valueName := "some"
	gaugeVal := 123.5
	counterVal := int64(123)
	gaugeValStr := strconv.FormatFloat(gaugeVal, 'g', -1, 64)
	counterValStr := strconv.FormatInt(counterVal, 10)

	ttPost := []struct {
		name          string
		url           string
		header        http.Header
		expStatusCode int
		expHeader     http.Header
	}{
		{
			name:          "wrong_type",
			url:           srv.URL + "/update/wrong_type/" + valueName + "/" + counterValStr,
			header:        header,
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "wrong_gauge_value_type",
			url:           srv.URL + "/update/gauge/" + valueName + "/lol",
			header:        header,
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "wrong_counter_value_type",
			url:           srv.URL + "/update/counter/" + valueName + "/lol",
			header:        header,
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "empty_param_name",
			url:           srv.URL + "/update/gauge//" + gaugeValStr,
			header:        header,
			expStatusCode: http.StatusNotFound,
		},
		{
			name:          "wrong_url",
			url:           srv.URL + "/update/",
			header:        header,
			expStatusCode: http.StatusNotFound,
		},
		{
			name:          "wrong_url_again",
			url:           srv.URL + "/",
			header:        header,
			expStatusCode: http.StatusNotFound,
		},
		{
			name:          "correct_gauge_data_int",
			url:           srv.URL + "/update/gauge/" + valueName + "/" + counterValStr,
			expStatusCode: http.StatusOK,
			header:        header,
			expHeader:     expHeader,
		},
		{
			name: "correct_gauge_data_float",
			url:  srv.URL + "/update/gauge/" + valueName + "/" + gaugeValStr,

			expStatusCode: http.StatusOK,
			header:        header,
			expHeader:     expHeader,
		},
		{
			name:          "correct_counter_data",
			url:           srv.URL + "/update/counter/" + valueName + "/" + counterValStr,
			expStatusCode: http.StatusOK,
			header:        header,
			expHeader:     expHeader,
		},
	}

	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	for _, tc := range ttPost {
		t.Run("Post_"+tc.name, func(t *testing.T) {
			resp, err := client.Post(tc.url, nil, tc.header)
			require.NoError(t, err)
			require.Equal(t, tc.expStatusCode, resp.StatusCode)
			if tc.expStatusCode == http.StatusOK {
				require.Equal(t, tc.expHeader.Get("Content-Type"), resp.Header.Get("Content-Type"))
				require.Equal(t, "0", resp.Header.Get("Content-Length"))
			}
			require.NoError(t, resp.Body.Close())
		})
	}

	ttGet := []struct {
		name          string
		url           string
		header        http.Header
		expStatusCode int
		expHeader     http.Header
		expData       string
	}{
		{
			name:          "unknown_name_counter",
			url:           srv.URL + "/value/counter/lol",
			header:        header,
			expStatusCode: http.StatusNotFound,
		},
		{
			name:          "unknown_name_gauge",
			url:           srv.URL + "/value/gauge/lol",
			header:        header,
			expStatusCode: http.StatusNotFound,
		},
		{
			name:          "unknown_type",
			url:           srv.URL + "/value/Z0zo/" + valueName,
			header:        header,
			expStatusCode: http.StatusBadRequest,
		},
		{
			name:          "correct_gauge_data",
			url:           srv.URL + "/value/gauge/" + valueName,
			header:        header,
			expStatusCode: http.StatusOK,
			expData:       gaugeValStr,
		},
		{
			name:          "correct_counter_data",
			url:           srv.URL + "/value/counter/" + valueName,
			header:        header,
			expStatusCode: http.StatusOK,
			expData:       counterValStr,
		},
	}

	for _, tc := range ttGet {
		t.Run("Get_"+tc.name, func(t *testing.T) {
			resp, err := client.Get(tc.url, tc.header)
			require.NoError(t, err)
			require.Equal(t, tc.expStatusCode, resp.StatusCode)
			if tc.expStatusCode == http.StatusOK {
				require.Equal(t, expHeader.Get("Content-Type"), resp.Header.Get("Content-Type"))
				buf, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NotZero(t, resp.Header.Get("Content-Length"))
				require.Equal(t, []byte(tc.expData), buf)
			}
			require.NoError(t, resp.Body.Close())
		})
	}
}

func Test_POST_GET(t *testing.T) {
	srv := httptest.NewServer(newRouter())
	defer srv.Close()
	// Create a new HTTP client with a default timeout
	timeout := 1000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	header := http.Header{
		"Content-Type": []string{"text/plain"},
	}

	expCounter := 0
	for i := range 3 {
		expCounter += i
		iStr := strconv.Itoa(i)
		expCounterStr := strconv.Itoa(expCounter)
		t.Run("Post_counter_"+iStr, func(t *testing.T) {
			resp, err := client.Post(srv.URL+"/update/counter/cntValName/"+iStr, nil, header)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})

		t.Run("Post_gauge_"+iStr, func(t *testing.T) {
			val := strconv.FormatFloat(float64(i), 'f', -1, 64)
			resp, err := client.Post(srv.URL+"/update/gauge/gaugeValName/"+val, nil, header)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.NoError(t, resp.Body.Close())
		})

		t.Run("Get_counter_"+iStr, func(t *testing.T) {
			resp, err := client.Get(srv.URL+"/value/counter/cntValName", header)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			buf, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, expCounterStr, string(buf))
			assert.NoError(t, resp.Body.Close())
		})

		t.Run("Get_gauge_"+iStr, func(t *testing.T) {
			resp, err := client.Get(srv.URL+"/value/gauge/gaugeValName", header)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			buf, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, iStr, string(buf))
			assert.NoError(t, resp.Body.Close())
		})
	}

	t.Run("Get_key_list", func(t *testing.T) {
		resp, err := client.Get(srv.URL+"/", header)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NoError(t, err)
		require.NotZero(t, resp.Header.Get("Content-Length"))
		require.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		buf, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		for _, key := range []string{"html", "Keys", "gaugevalname", "cntvalname"} {
			assert.Contains(t, string(buf), key)
		}
		require.NoError(t, resp.Body.Close())
	})
}
