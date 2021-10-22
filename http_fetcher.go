package main

import (
	"bufio"
	"fmt"
	"github.com/phuslu/log"
	"gorobei/utils"
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

type (
	HttpFetcher interface {
		FetchHtml(url string) (string, error)
		FetchImage(url string) (string, error)
	}

	httpFetcherImpl struct {

	}
)

var ImageExt = map[string]string{
	"image/bmp":     "bmp",
	"image/gif":     "gif",
	"image/jpeg":    "jpeg",
	"image/png":     "png",
	"image/svg+xml": "svg",
	"image/tiff":    "tiff",
	"image/webp":    "webp",
}


func (f *httpFetcherImpl) FetchHtml(url string) (string, error) {
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ctype := resp.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(ctype)
	if err != nil {
		return "", fmt.Errorf("cannot parse `Content-Type`=`%v`. %v", ctype)
	}



	if resp.StatusCode != http.StatusOK {
		return "", httpResponseError(resp, mediatype)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err

}

func (f *httpFetcherImpl) FetchImage(url string) (string, error) {
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mediatype string
	mediatype, _, err = mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", httpResponseError(resp, mediatype)
	}

	ext, ok := ImageExt[mediatype]
	if !ok {
		return "", fmt.Errorf("unsupported mediatype: %s", mediatype)
	}
	tf, err := ioutil.TempFile("", "*."+ext)
	defer tf.Close()
	log.Info().Str("file", tf.Name()).Msg("image temp file name")
	r := bufio.NewReader(resp.Body)
	_, err = r.WriteTo(tf)
	if err != nil {
		return "", err
	}
	return tf.Name(), nil
}

func httpResponseError(resp *http.Response, mediatype string) error {
	//var mediatype string
	//ctype := resp.Header.Get("Content-Type")
	//mediatype, _, err := mime.ParseMediaType(ctype)
	//if err != nil {
	//	return fmt.Errorf("httpErrorText: cannot parse `Content-Type`=`%v`. %v", ctype)
	//}

	text := ""
	switch mediatype {
	case "text/plain", "text/html", "application/json":
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			text = err.Error()
		} else {
			text = string(body)
		}
	}
	return fmt.Errorf("http error %v:\n%v", resp.StatusCode, utils.FirstN(text,300))
}