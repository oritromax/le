package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sjsakib/gudam/pkg/utils"
	"github.com/sjsakib/gudam/pkg/nanoid"
)

type fileHandler struct {
	defaultServer http.Handler
	root          http.Dir
}

func newHandler(dir string) http.Handler {
	return &fileHandler{
		defaultServer: http.FileServer(http.Dir(dir)),
		root:          http.Dir(dir),
	}
}

func (h fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(string(h.root), r.URL.Path)

	connID := nanoid.New()
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	log.Printf("[REQUEST] %s | %s - %s", connID, clientIP, r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Printf("[405] %s | %s - %s", connID, clientIP, r.URL.Path)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			log.Printf("[404] %s | %s - %s", connID, clientIP, r.URL.Path)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("[500] %s | %s - error stating %s: %v", connID, clientIP, r.URL.Path, err)
		return
	}

	if info.IsDir() {
		log.Printf("Using default file server for %s | %s - %s", connID, clientIP, r.URL.Path)
		h.defaultServer.ServeHTTP(w, r)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Printf("[500] %s | %s - error opening %s: %v", connID, clientIP, r.URL.Path, err)
		return
	}
	defer file.Close()

	var reader io.Reader = file
	var contentLength int64 = info.Size()
	var startByte int64 = 0
	var endByte int64 = 0
	var transferStart = time.Now()

	rng := r.Header.Get("Range")
	if rng != "" {
		startByte, endByte, err = utils.ParseRangeHeader(rng, info.Size())
		if err != nil {
			http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
			log.Printf("[416] %s | %s - invalid range %s for %s: %v", connID, clientIP, rng, r.URL.Path, err)
			return
		}

		if _, err := file.Seek(startByte, io.SeekStart); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("[500] %s | %s - error seeking %s: %v", connID, clientIP, r.URL.Path, err)
			return
		}

		contentLength = endByte - startByte + 1
		reader = io.LimitReader(file, contentLength)

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startByte, endByte, info.Size()))
		w.WriteHeader(http.StatusPartialContent)

		log.Printf("[206] %s - %s range %d-%d", clientIP, r.URL.Path, startByte, endByte)
	} else {
		log.Printf("[200] %s - %s (%d bytes)", clientIP, r.URL.Path, info.Size())
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("ETag", fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size()))

	fileName := filepath.Base(path)
	var totalSent int64 = 0
	var totalMBSent float64
	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		bufferStart := time.Now()

		n, readErr := reader.Read(buf)
		if readErr != nil {
			if readErr != io.EOF {
				log.Printf("[XFER ERROR] %s | %s - %s: %v", connID, clientIP, fileName, readErr)
			}
			break
		}

		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				log.Printf("[XFER INTERRUPTED] %s | %s | %s after %.2fMB", connID, clientIP, fileName, totalMBSent)
				break
			}
			totalSent += int64(n)
			totalMBSent = float64(totalSent) / 1024 / 1024
			mbps := 1 / time.Since(bufferStart).Seconds()
			progress := float64(totalSent) / float64(info.Size()) * 100
			log.Printf("[XFER] %s | %s - %s: %.2fMB sent, %.2fMB/s, %.2f%%", connID, clientIP, fileName, totalMBSent, mbps, progress)
		}
	}

	log.Printf("[DONE] %s | %s - %s: %.2fMB in %s", connID, clientIP, fileName, totalMBSent, time.Since(transferStart))
}
