package model

import (
	"fmt"
	"net"
	"strings"
)

type AggregatedNicError struct {
	errorMap map[net.Addr]error
}

func NewAggregatedNicError(errorMap map[net.Addr]error) *AggregatedNicError {
	if len(errorMap) <= 0 {
		return nil
	}
	return &AggregatedNicError{errorMap}
}

func (a *AggregatedNicError) GetErrors() map[net.Addr]error {
	clone := make(map[net.Addr]error, len(a.errorMap))
	for addr, err := range a.errorMap {
		clone[addr] = err
	}
	return clone
}

func (a *AggregatedNicError) GetError(addr net.Addr) error {
	return a.errorMap[addr]
}

func (a *AggregatedNicError) Error() string {
	buffer := make([]string, len(a.errorMap))
	index := 0
	for addr, err := range a.errorMap {
		buffer[index] = fmt.Sprintf("%s on %s", err, addr.String())
		index++
	}
	return strings.Join(buffer, "\n")
}
