package clf

import (
	"bytes"
	"fmt"
	"io"
)

type Request struct {
	Method   string
	Resource string
	Protocol string
}

func (r Request) String() string {
	return fmt.Sprintf("%s %s %s",
		r.Method,
		r.Resource,
		r.Protocol,
	)
}

type requestState int

const (
	stateRequestBegin requestState = iota
	stateRequestMethod
	stateRequestResource
	stateRequestProtocol
	stateRequestEnd
)

func ParseRequest(val []byte) (r Request, err error) {
	buf := bytes.NewBuffer(val)
	state := stateRequestBegin

	var field []byte
	for state != stateRequestEnd {
		if err == io.EOF {
			err = nil
			break
		}

		if err != nil {
			return
		}

		switch state {
		case stateRequestMethod:
			if field, err = buf.ReadBytes(' '); err == nil {
				field = field[:len(field)-1]
				r.Method = string(field)
			}
		case stateRequestResource:
			if field, err = buf.ReadBytes(' '); err == nil {
				field = field[:len(field)-1]
				r.Resource = string(field)
			}
		case stateRequestProtocol:
			r.Protocol = buf.String()
		}

		state++
	}

	return
}
