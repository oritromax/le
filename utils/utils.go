package utils

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
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

	fmt.Printf("Matches: %v | header: %q\n", matches, header)

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
