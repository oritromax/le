package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "reqId"
)

func GetLocalIP() (string, error) {
	// Connect to a dummy address; doesn't have to be reachable
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func ParseRangeHeader(header string, size int64) (start, end int64, err error) {
	if header == "" {
		return 0, 0, nil
	}
	re := regexp.MustCompile(`^bytes=(\d*)-(\d*)$`)

	matches := re.FindStringSubmatch(header)

	errMsg := fmt.Errorf("invalid range header: %q for size %d", header, size)

	if len(matches) < 3 || (matches[1] == "" && matches[2] == "") {
		return 0, 0, errMsg
	}

	if matches[1] != "" {
		start, _ = strconv.ParseInt(matches[1], 10, 64)
	}

	if matches[2] != "" {
		end, _ = strconv.ParseInt(matches[2], 10, 64)
	}

	if matches[1] == "" {
		start = size - end
		end = size - 1
	}

	if matches[2] == "" {
		end = size - 1
	}

	if start > end || end > size {
		return 0, 0, errMsg
	}

	return start, end, nil
}

var ErrForbiddenPath = errors.New("forbidden path")

// SecureJoin ensures that the joined path is within the base directory
func SecureJoin(base, path string) (string, error) {
	// get root path
	absRoot, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	// because macOS is a special snowflake
	absRoot, err = filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", err
	}
	absRoot = filepath.Clean(absRoot)

	targetPath := filepath.Join(absRoot, path)

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	parentPath := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil && os.IsNotExist(err) {
		evalParent, err := filepath.EvalSymlinks(parentPath)
		if err == nil {
			absPath = filepath.Join(evalParent, fileName)
		}
	} else if err != nil {
		return "", err
	}

	absPath = filepath.Clean(absPath)

	// prevent prefix matching for path traversal
	if !strings.HasPrefix(absPath, absRoot+(string(filepath.Separator))) && absPath != absRoot {
		return "", ErrForbiddenPath
	}

	return absPath, nil
}
