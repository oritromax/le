package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.sakib.dev/le/logger"
	"go.sakib.dev/le/pkg/nanoid"
	"go.sakib.dev/le/pkg/utils"
)

type handler struct {
	defaultServer http.Handler
	root          http.Dir
}

func newHandler(dir string) http.Handler {
	return &handler{
		defaultServer: http.FileServer(http.Dir(dir)),
		root:          http.Dir(dir),
	}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqHelper := newReqHelper(w, r)

	reqHelper.attachReqId()

	reqHelper.logRequest()

	if r.Method != http.MethodGet {
		reqHelper.error("Method Not Allowed", nil, http.StatusMethodNotAllowed)
		return
	}

	absPath, err := utils.SecureJoin(string(h.root), r.URL.Path)

	if errors.Is(err, utils.ErrForbiddenPath) {
		reqHelper.error("FORBIDDEN", err, http.StatusForbidden)
		return
	} else if err != nil {
		reqHelper.internalServerError(err)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			reqHelper.error("NOT FOUND", err, http.StatusNotFound)
			return
		}
		reqHelper.internalServerError(err)
		return
	}

	if info.IsDir() {
		// check if request is coming from a browser
		acceptHeader := r.Header.Get("Accept")
		isBrowser := strings.Contains(acceptHeader, "text/html")

		if isBrowser {
			slog.InfoContext(reqHelper.ctx, "OK - Serving directory with pretty UI", "path", r.URL.Path)
			h.serveDirectory(w, r, absPath)
		} else {
			slog.InfoContext(reqHelper.ctx, "OK - Serving directory default file server", "path", r.URL.Path)
			h.defaultServer.ServeHTTP(w, r)
		}
		return
	}

	file, err := os.Open(absPath)
	if err != nil {
		reqHelper.internalServerError(err)
		return
	}
	defer file.Close()
	var transferStart = time.Now()

	contentLength, reader, err := reqHelper.handleRange(file, info)
	if err != nil {
		if errors.Is(err, ErrInvalidRangeHeader) {
			reqHelper.error("Invalid Range", err, http.StatusRequestedRangeNotSatisfiable)
			return
		}
		reqHelper.internalServerError(err)
		return

	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("ETag", fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size()))

	fileName := filepath.Base(absPath)
	var totalSent int64 = 0
	var totalMBSent float64
	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		bufferStart := time.Now()

		n, readErr := reader.Read(buf)
		if readErr != nil {
			if readErr != io.EOF {
				slog.ErrorContext(reqHelper.ctx, "Error reading file", "error", readErr, "file", fileName)
			}
			break
		}

		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				slog.ErrorContext(reqHelper.ctx, "Error writing response", "error", writeErr, "file", fileName)
				break
			}
			totalSent += int64(n)
			totalMBSent = float64(totalSent) / 1024 / 1024
			mbps := 1 / time.Since(bufferStart).Seconds()
			progress := float64(totalSent) / float64(info.Size()) * 100
			slog.InfoContext(reqHelper.ctx, "XFER", "sent_mb", totalMBSent, "speed_mbps", mbps, "progress", progress, "file", fileName)
		}
	}

	slog.InfoContext(reqHelper.ctx, "TRANSFER COMPLETE", "file", fileName, "totalSent_mb", totalMBSent, "duration", time.Since(transferStart))
}

type reqHelper struct {
	w   http.ResponseWriter
	r   *http.Request
	ctx context.Context
}

func newReqHelper(w http.ResponseWriter, r *http.Request) *reqHelper {
	return &reqHelper{
		w:   w,
		r:   r,
		ctx: r.Context(),
	}
}

func (h *reqHelper) attachReqId() *context.Context {
	reqId := nanoid.New()
	ctx := context.WithValue(h.ctx, utils.RequestIDKey, reqId)
	h.r = h.r.WithContext(ctx)
	h.ctx = ctx
	return &ctx
}

func (h reqHelper) logRequest() {
	clientIP, _, _ := net.SplitHostPort(h.r.RemoteAddr)
	slog.InfoContext(h.ctx, "REQUEST",
		"clientIP", clientIP,
		"method", h.r.Method,
		"path", h.r.URL.Path)

}

var ErrInvalidRangeHeader = errors.New("invalid range header")

func (h *reqHelper) handleRange(file *os.File, fileInfo os.FileInfo) (int64, io.Reader, error) {
	rng := h.r.Header.Get("Range")
	var contentLength = fileInfo.Size()
	var reader io.Reader = file
	if rng != "" {
		startByte, endByte, err := utils.ParseRangeHeader(rng, fileInfo.Size())
		if err != nil {
			return 0, nil, ErrInvalidRangeHeader
		}

		if _, err := file.Seek(startByte, io.SeekStart); err != nil {
			return 0, nil, err
		}

		contentLength = endByte - startByte + 1
		reader = io.LimitReader(file, contentLength)

		h.w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startByte, endByte, fileInfo.Size()))
		h.w.WriteHeader(http.StatusPartialContent)

		slog.InfoContext(h.ctx, "PARTIAL", "path", h.r.URL.Path, "start", startByte, "end", endByte, "total", contentLength, logger.StatusCodeKey, http.StatusPartialContent)
	} else {
		slog.InfoContext(h.ctx, "OK", "path", h.r.URL.Path, "size", contentLength, logger.StatusCodeKey, http.StatusOK)
	}

	return contentLength, reader, nil
}

func (h *reqHelper) internalServerError(err error) {
	h.error("Internal Server Error", err, http.StatusInternalServerError)
}

func (h *reqHelper) error(mgs string, err error, statusCode int) {
	http.Error(h.w, mgs, statusCode)
	slog.ErrorContext(h.ctx, "", logger.StatusCodeKey, statusCode, "error", err)
}
