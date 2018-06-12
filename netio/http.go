package netio

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// HTTPWriter gives io.Write methods to a http Client
type HTTPWriter struct {
	Client *http.Client
	Method string
	URL    string
	Header http.Header
}

func (w *HTTPWriter) Write(p []byte) (int, error) {
	br := bytes.NewReader(p)
	req, err := http.NewRequest(w.Method, w.URL, br)
	if err != nil {
		return 0, err
	}

	req.Header = w.Header

	res, err := w.Client.Do(req)
	if err != nil {
		return 0, err
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
	return len(p), nil
}
