package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type harentry struct {
	Request struct {
		URL string
	}
	Response struct {
		Status  int
		Content struct {
			Text     string
			Encoding string
			MimeType string
		}
	}
}

type harlog struct {
	Entries []harentry
}

type har struct {
	Log harlog
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: decode.go <harfile>\n\n")
		os.Exit(1)
	}

	var fname = os.Args[1]
	var f, err = os.Open(fname)
	if err != nil {
		log.Println("Couldn't open %q: %s", fname, err)
	}

	var data []byte
	data, err = ioutil.ReadAll(f)
	if err != nil {
		log.Println("Couldn't read %q: %s", fname, err)
	}

	var h har
	json.Unmarshal(data, &h)

	for _, e := range h.Log.Entries {
		if e.Response.Status != 200 {
			continue
		}

		var fixedURL = strings.Replace(e.Request.URL, "%2F", "/", -1)
		var filename = filepath.Base(fixedURL)
		var ext = filepath.Ext(e.Request.URL)
		if ext == "" {
			var exts []string
			exts, err = mime.ExtensionsByType(e.Response.Content.MimeType)
			if err != nil {
				log.Printf("Unable to process %q's mime type %q: %s", e.Request.URL, e.Response.Content.MimeType, err)
				continue
			}
			if len(exts) > 0 {
				filename += exts[0]
			}
		}
		log.Printf("Processing %q", filename)

		var content = []byte(e.Response.Content.Text)
		if e.Response.Content.Encoding == "base64" {
			var dbuf = make([]byte, base64.StdEncoding.DecodedLen(len(content)))
			_, err = base64.StdEncoding.Decode(dbuf, content)
			if err != nil {
				log.Printf("Error decoding encoded content for %q: %s", filename, err)
				os.Exit(1)
			}
		}

		os.Mkdir("./output", 0750)
		var out *os.File
		out, err = os.Create(filepath.Join("output", filename))
		if err != nil {
			log.Printf("Error creating file for %q: %s", filename, err)
			os.Exit(1)
		}
		defer out.Close()
		_, err = out.Write(content)
		if err != nil {
			log.Printf("Error writing %q: %s", filename, err)
			os.Exit(1)
		}
	}
}
