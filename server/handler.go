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

	"github.com/sjsakib/gudam/utils"
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

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Printf("[405] %s - %s", r.RemoteAddr, r.URL.Path)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			log.Printf("[404] %s - %s", r.RemoteAddr, r.URL.Path)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("[500] %s - error stating %s: %v", r.RemoteAddr, r.URL.Path, err)
		return
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if info.IsDir() {
		log.Printf("Using default file server for %s - %s", clientIP, r.URL.Path)
		h.defaultServer.ServeHTTP(w, r)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Printf("[500] %s - error opening %s: %v", clientIP, r.URL.Path, err)
		return
	}
	defer file.Close()

	rng := r.Header.Get("Range")
	var reader io.Reader = file
	var contentLength int64 = info.Size()
	var logMsg string
	var total int64
	var transferStart = time.Now()
	if rng != "" {
		start, end, err := utils.ParseRangeHeader(rng, info.Size())
		if err != nil {
			http.Error(w, "Invalid Range", http.StatusRequestedRangeNotSatisfiable)
			log.Printf("[416] %s - invalid range %s for %s: %v", clientIP, rng, r.URL.Path, err)
			return
		}
		if _, err := file.Seek(start, io.SeekStart); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("[500] %s - error seeking %s: %v", clientIP, r.URL.Path, err)
			return
		}
		reader = io.LimitReader(file, end-start+1)
		contentLength = end - start + 1
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
		w.WriteHeader(http.StatusPartialContent)
		logMsg = fmt.Sprintf("[206] %s - %s range %d-%d", clientIP, r.URL.Path, start, end)
	} else {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		logMsg = fmt.Sprintf("[200] %s - %s (%d bytes)", clientIP, r.URL.Path, info.Size())
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size()))

	log.Print(logMsg)

	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				log.Printf("[XFER INTERRUPTED] %s - %s after %.2fMB", clientIP, r.URL.Path, float64(total)/1024/1024)
				break
			}
			total += int64(n)
			log.Printf("[XFER] %s - %s: %.2fMB sent", clientIP, r.URL.Path, float64(total)/1024/1024)
		}
		if readErr != nil {
			if readErr != io.EOF {
				log.Printf("[XFER ERROR] %s - %s: %v", clientIP, r.URL.Path, readErr)
			}
			break
		}
	}

	log.Printf("[DONE] %s - %s: %.2fMB in %s", clientIP, r.URL.Path, float64(total)/1024/1024, time.Since(transferStart))
}
