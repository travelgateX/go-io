package httpio

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// HTTP gives io.Write methods to a http Client
type HTTP struct {
	Client *http.Client
	Method string
	URL    string
	Header http.Header
}

func (w *HTTP) Write(p []byte) (int, error) {
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
