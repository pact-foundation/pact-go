// Package utils contains a number of helper / utility functions.
package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// GetFreePort Gets an available port by asking the kernal for a random port
// ready and available for use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// FindPortInRange Iterate through CSV or Range of ports to find open port
// Valid inputs are "8081", "8081,8085", "8081-8085". Do not combine
// list and range
func FindPortInRange(s string) (int, error) {
	// Take care of csv and single value
	if !strings.Contains(s, "-") {
		ports := strings.Split(strings.TrimSpace(s), ",")
		for _, p := range ports {
			i, err := strconv.Atoi(p)
			if err != nil {
				return 0, err
			}
			err = checkPort(i)
			if err != nil {
				continue
			}
			return i, nil
		}
		return 0, errors.New("all passed ports are unusable")
	}
	// Now take care of ranges
	ports := strings.Split(strings.TrimSpace(s), "-")
	if len(ports) != 2 {
		return 0, errors.New("invalid range passed")
	}
	lower, err := strconv.Atoi(ports[0])
	if err != nil {
		return 0, err
	}
	upper, err := strconv.Atoi(ports[1])
	if err != nil {
		return 0, err
	}
	if upper < lower {
		return 0, errors.New("invalid range passed")
	}
	for i := lower; i <= upper; i++ {
		err = checkPort(i)
		if err != nil {
			continue
		}
		return i, nil
	}
	return 0, errors.New("all passed ports are unusable")
}

func checkPort(p int) error {
	s := fmt.Sprintf("localhost:%d", p)
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	return nil
}
