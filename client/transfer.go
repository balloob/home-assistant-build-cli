package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (c *RestClient) transferRequest(ctx context.Context, method, endpoint string, query url.Values, body io.Reader, contentType string) (*http.Response, error) {
	u, err := url.Parse(strings.TrimRight(c.BaseURL, "/") + "/api/" + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid Home Assistant URL")
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("cannot create transfer request")
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig(c.VerifySSL)
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: c.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		// URL errors can contain the backup decryption password.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("file transfer failed: %w", err)
	}
	return resp, nil
}

func transferStatusError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	code := ErrCodeAPIError
	switch resp.StatusCode {
	case 401:
		code = ErrCodeAuthenticationError
	case 403:
		code = ErrCodePermissionDenied
	case 404:
		code = ErrCodeNotFound
	case 400, 422:
		code = ErrCodeValidationError
	}
	// Server bodies can echo credentials or upload contents.
	return &APIError{Code: code, Message: fmt.Sprintf("File transfer returned HTTP %d", resp.StatusCode), StatusCode: resp.StatusCode, Transport: "rest"}
}

func (c *RestClient) DownloadFile(ctx context.Context, endpoint string, query url.Values, destination string) (size int64, err error) {
	if _, statErr := os.Lstat(destination); statErr == nil {
		return 0, fmt.Errorf("output file already exists")
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return 0, statErr
	}
	file, err := os.CreateTemp(filepath.Dir(destination), ".hab-download-*")
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, os.Remove(file.Name())) }()
	size, err = c.downloadToFile(ctx, endpoint, query, file)
	if err != nil {
		return 0, err
	}
	// A hard link publishes the complete file without overwriting a concurrent writer.
	if err = os.Link(file.Name(), destination); err != nil {
		return 0, err
	}
	return size, nil
}

func (c *RestClient) downloadToFile(ctx context.Context, endpoint string, query url.Values, file *os.File) (size int64, err error) {
	defer func() { err = errors.Join(err, file.Close()) }()
	resp, err := c.transferRequest(ctx, http.MethodGet, endpoint, query, nil, "")
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, resp.Body.Close()) }()
	if err = transferStatusError(resp); err != nil {
		return 0, err
	}
	return io.Copy(file, resp.Body)
}

func (c *RestClient) UploadFile(ctx context.Context, endpoint string, query url.Values, filename string, fields map[string]string) (result any, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("upload requires a regular file")
	}
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	done := make(chan error, 1)
	go func() {
		writeErr := writeUpload(multipartWriter, file, fields)
		closeErr := writer.CloseWithError(writeErr)
		done <- errors.Join(writeErr, closeErr)
	}()
	resp, requestErr := c.transferRequest(ctx, http.MethodPost, endpoint, query, reader, multipartWriter.FormDataContentType())
	// A server can reject the request before consuming the file.
	closeErr := reader.Close()
	writeErr := <-done
	if requestErr != nil {
		return nil, errors.Join(requestErr, closeErr)
	}
	defer func() { err = errors.Join(err, resp.Body.Close()) }()
	if err = transferStatusError(resp); err != nil {
		return nil, err
	}
	if err = errors.Join(writeErr, closeErr); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 4<<20))
	decoder.UseNumber()
	if err = decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid upload response: %w", err)
	}
	return result, nil
}

func writeUpload(writer *multipart.Writer, file *os.File, fields map[string]string) error {
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	contentType := mime.TypeByExtension(filepath.Ext(file.Name()))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filepath.Base(file.Name())}))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	return writer.Close()
}
