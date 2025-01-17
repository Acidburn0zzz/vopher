package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
)

// HEAD url, find out "Content-Disposition: attachment; filename=foo.EXT"
func httpdetectFtype(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	content := resp.Header.Get("Content-Disposition")
	if content == "" {
		return "", fmt.Errorf("can't detect filetype of %q", url)
	}

	const fn = "filename="
	idx := strings.Index(content, fn)
	if idx == -1 || len(content) == idx+len(fn) {
		return "", fmt.Errorf("invalid 'Content-Disposition' header for %q", url)
	}

	return path.Ext(content[idx+len(fn):]), nil
}

func httpGET(w io.Writer, url, checkSha1 string) (err error) {

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(resp)
		return fmt.Errorf("%d for %q", resp.StatusCode, url)
	}

	reader := io.Reader(resp.Body)
	hasher := sha1.New()

	if checkSha1 != "" {
		reader = io.TeeReader(reader, hasher)
	}

	/*
		// NOTE: idea to report sub-progress, but maybe not worth the
		// effort since plugins are really small

		if resp.ContentLength > 0 {
			progress := NewProgressTicker(resp.ContentLength)
			defer progress.Stop()
			go progress.Start(out, 2*time.Millisecond)
			reader = io.TeeReader(reader, progress)
		}
	*/

	if _, err := io.Copy(w, reader); err != nil {
		return err
	}

	sha1Sum := hex.EncodeToString(hasher.Sum(nil))
	if checkSha1 != "" && checkSha1 != sha1Sum {
		return fmt.Errorf("sha1 does not match: got %s, expected %s", sha1Sum, checkSha1)
	}

	return err
}
