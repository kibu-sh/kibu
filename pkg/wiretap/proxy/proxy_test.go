package proxy

//func TestProxy(t *testing.T) {
//	capture := NewChannelBasedCaptureTransport(results).
//		WithTimeout(time.Second * 4)
//
//	proxyServer := httptest.NewServer(
//		NewCaptureProxy(capture),
//	)
//	t.Cleanup(func() {
//		proxyServer.Close()
//	})
//
//	echoServer := httptest.NewServer(EchoHandler())
//	t.Cleanup(func() {
//		echoServer.Close()
//	})
//
//	proxyClient := proxyClientFromServer(proxyServer)
//	t.Run("should successfully proxy a request", func(t *testing.T) {
//		res, err := proxyClient.Get(echoServer.URL)
//		require.NoError(t, err)
//
//		var body map[string]string
//		err = json.NewDecoder(res.Body).Decode(&body)
//		require.NoError(t, err)
//		require.Equal(t, "hello world", body["message"])
//
//		_, err = readSnapshotWithTimeout(results, time.Second*3)
//		require.NoError(t, err)
//	})
//
//	t.Run("should successfully capture a round trip record", func(t *testing.T) {
//		_, err := proxyClient.Do(&http.Request{
//			Method: http.MethodPost,
//			URL:    lo.Must(url.Parse(echoServer.URL)),
//			Header: http.Header{
//				"Content-Type": []string{"application/json"},
//				"X-Test":       []string{"true"},
//			},
//			Body: io.NopCloser(strings.NewReader(`{"hello": "echo"}`)),
//		})
//		require.NoError(t, err)
//
//		record, err := readSnapshotWithTimeout(results, time.Second*3)
//		require.NoError(t, err)
//		require.Equal(t, "application/json", record.Response.Header.Get("Content-Type"))
//		require.Equal(t, http.StatusOK, record.Response.StatusCode)
//		require.Equal(t, http.MethodPost, record.Request.Method)
//		require.Equal(t, "true", record.Request.Header.Get("X-Test"))
//		require.Equal(t, "application/json", record.Request.Header.Get("Content-Type"))
//
//		parsedRequestURL := lo.Must(url.Parse(record.Request.URL.String()))
//		parsedEchoServerURL := lo.Must(url.Parse(echoServer.URL))
//		require.Equal(t, parsedEchoServerURL.Host, parsedRequestURL.Host)
//		require.Equal(t, "/", parsedRequestURL.Path)
//
//		var body map[string]string
//		err = json.NewDecoder(record.Request.Body).Decode(&body)
//		require.NoError(t, err)
//		require.Equal(t, "echo", body["hello"])
//	})
//
//	t.Run("should successfully record a proxy failure", func(t *testing.T) {
//		res, err := proxyClient.Do(&http.Request{
//			Method: http.MethodGet,
//			URL:    lo.Must(url.Parse(echoServer.URL)),
//			Header: http.Header{
//				ScenarioHeader: []string{ScenarioDisconnect},
//			},
//		})
//		require.NoError(t, err)
//		require.NotNil(t, res)
//
//		record, err := readSnapshotWithTimeout(results, time.Second*3)
//		require.NoError(t, err)
//		require.Equal(t, http.StatusBadGateway, record.Response.StatusCode)
//	})
//
//	t.Run("should successfully record a proxy timeout", func(t *testing.T) {
//		res, err := proxyClient.Do(&http.Request{
//			Method: http.MethodGet,
//			URL:    lo.Must(url.Parse(echoServer.URL)),
//			Header: http.Header{
//				ScenarioHeader: []string{ScenarioTimeout},
//			},
//		})
//		require.NoError(t, err)
//		require.NotNil(t, res)
//
//		record, err := readSnapshotWithTimeout(results, time.Second*3)
//		require.NoError(t, err)
//		require.Equal(t, http.StatusGatewayTimeout, record.Response.StatusCode)
//	})
//}
