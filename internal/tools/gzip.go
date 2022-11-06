package tools

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gzipWriter gzipWriter) Write(b []byte) (int, error) {
	return gzipWriter.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// gzip Decode
		if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(writer, request)
			return
		}

		if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(request.Body)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				_ = gzipReader.Close()
			}()

			body, err := io.ReadAll(gzipReader)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}

			request.Body = io.NopCloser(bytes.NewReader(body))
			request.ContentLength = int64(len(body))
		}

		gzipReader, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
		if err != nil {
			fmt.Println(err)
			if _, err := io.WriteString(writer, err.Error()); err != nil {
				fmt.Println(err)
				return
			}
			return
		}
		defer func() {
			_ = gzipReader.Close()
		}()

		writer.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: writer, Writer: gzipReader}, request)
	})
}
