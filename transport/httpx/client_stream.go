package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// InvokeReader uploads body without encoding or buffering it, then decodes the
// response normally. If body is an io.ReadCloser, ownership transfers on entry,
// including validation errors. Set an appropriate Content-Type via call options.
// Middleware receives body and must not retry or concurrently replay the call.
func (c *Client) InvokeReader(ctx context.Context, method, path string, body io.Reader, reply interface{}, opts ...CallOption) error {
	if body != nil && isNilCoderOptionValue(body) {
		return errors.New("httpx: typed nil upload reader")
	}
	if closer, ok := body.(io.ReadCloser); ok {
		owned := closeOnce(closer)
		body = owned
		defer owned.Close()
	}
	call, err := parseCallOptions(path, opts)
	if err != nil {
		return err
	}
	ctx, req, err := c.prepareInvocation(ctx, method, path, body, call)
	if err != nil {
		return err
	}
	_, err = c.execute(ctx, req, body, call, responseHandler{
		consume: func(ctx context.Context, res *http.Response) (interface{}, error) {
			if err := c.decoder(ctx, res, reply); err != nil {
				return nil, err
			}
			return reply, nil
		},
	}, true)
	return err
}

type DownloadInfo struct {
	StatusCode int
	Header     http.Header
	Bytes      int64 // bytes successfully written, including partial results on error
}

// UnexpectedStatusError describes a final non-2xx download response. Error
// bodies with status >=400 normally use the configured decoder's error instead.
type UnexpectedStatusError struct {
	StatusCode int
	Location   string
}

func (e *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("httpx: unexpected download status %d", e.StatusCode)
}

// DownloadMaxBytes limits bytes written by Download. Zero (the default) means
// unlimited. Error response bodies use the response decoder's separate limit.
func DownloadMaxBytes(n int64) CallOption {
	return func(c *callInfo) {
		if n < 0 {
			c.configErr = errors.New("httpx: download limit must not be negative")
			return
		}
		c.downloadMaxBytes = n
	}
}

// Download performs a GET and copies a successful response to dst. It closes
// the response body, never closes dst, and does not write error pages to dst.
// The client's existing total timeout still applies; use WithTimeout(0) with
// a caller deadline for long transfers. Blocking custom writers must cooperate
// with cancellation. Middleware spans the copy and receives nil as its input.
func (c *Client) Download(ctx context.Context, path string, dst io.Writer, opts ...CallOption) (info DownloadInfo, err error) {
	if isNilCoderOptionValue(dst) {
		return info, errors.New("httpx: download writer is required")
	}
	call, err := parseCallOptions(path, opts)
	if err != nil {
		return info, err
	}
	ctx, req, err := c.prepareInvocation(ctx, http.MethodGet, path, nil, call)
	if err != nil {
		return info, err
	}
	_, err = c.execute(ctx, req, nil, call, responseHandler{
		received: func(res *http.Response) {
			info.StatusCode, info.Header = res.StatusCode, res.Header.Clone()
		},
		consume: func(ctx context.Context, res *http.Response) (interface{}, error) {
			if res.StatusCode < 200 || res.StatusCode >= 300 {
				if res.StatusCode >= 400 {
					if err := c.decoder(ctx, res, nil); err != nil {
						return info, err
					}
				}
				return info, &UnexpectedStatusError{StatusCode: res.StatusCode, Location: res.Header.Get("Location")}
			}
			if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusResetContent {
				return info, nil
			}
			if call.downloadMaxBytes > 0 && res.ContentLength > call.downloadMaxBytes {
				return info, &ResponseTooLargeError{Limit: call.downloadMaxBytes, StatusCode: res.StatusCode}
			}
			var copyErr error
			info.Bytes, copyErr = copyDownload(ctx, dst, res.Body, call.downloadMaxBytes, res.StatusCode)
			return info, copyErr
		},
	}, true)
	return info, err
}

func copyDownload(ctx context.Context, dst io.Writer, src io.Reader, limit int64, status int) (written int64, err error) {
	buf := make([]byte, 32<<10)
	emptyReads := 0
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		p := buf
		atLimit := limit > 0 && written == limit
		if atLimit {
			p = buf[:1] // Probe for excess; never write the extra byte.
		} else if limit > 0 && int64(len(p)) > limit-written {
			p = p[:limit-written]
		}
		n, readErr := src.Read(p)
		if n < 0 || n > len(p) {
			return written, errors.New("httpx: invalid reader count")
		}
		if atLimit && n > 0 {
			return written, &ResponseTooLargeError{Limit: limit, StatusCode: status}
		}
		if n > 0 {
			emptyReads = 0
			if err := ctx.Err(); err != nil {
				return written, err
			}
			nw, writeErr := dst.Write(p[:n])
			if nw < 0 || nw > n {
				return written, errors.New("httpx: invalid writer count")
			}
			written += int64(nw)
			if writeErr != nil {
				return written, writeErr
			}
			if nw != n {
				return written, io.ErrShortWrite
			}
		} else if readErr == nil {
			emptyReads++
			if emptyReads >= 100 {
				return written, io.ErrNoProgress
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return written, nil
			}
			return written, readErr
		}
	}
}
