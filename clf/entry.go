package clf

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Entry struct {
	IP         net.IP
	Identity   string
	UserID     string
	Time       time.Time
	Request    Request
	StatusCode int
	Size       int
}

func (e Entry) String() string {
	identity := "-"
	if e.Identity != "" {
		identity = e.Identity
	}

	userID := "-"
	if e.UserID != "" {
		userID = e.UserID
	}

	size := "-"
	if e.Size > 0 {
		size = strconv.FormatInt(int64(e.Size), 10)
	}

	return fmt.Sprintf(`%s %s %s [%s] "%s" %d %s`,
		e.IP,
		identity,
		userID,
		e.Time.Format(layout),
		e.Request,
		e.StatusCode,
		size,
	)
}

type lineState int

const (
	stateLineBegin lineState = iota
	stateLineIP
	stateLineIdentity
	stateLineUserID
	stateLineTimeBegin
	stateLineTime
	stateLineRequestBegin
	stateLineRequest
	stateLineRequestEnd
	stateLineStatusCode
	stateLineSize
	stateLineEnd
)

const layout = "02/Jan/2006:15:04:05 -0700"

func ParseEntry(line []byte) (e Entry, err error) {
	var unavail = []byte("-") // can't make this a const

	buf := bytes.NewBuffer(line)
	state := stateLineIP

	var field []byte
	for state != stateLineEnd {
		if err == io.EOF {
			err = nil
			break
		}

		if err != nil {
			return
		}

		switch state {
		case stateLineIP:
			if field, err = buf.ReadBytes(' '); err == nil {
				field = field[:len(field)-1]
				e.IP = net.ParseIP(string(field))
			}
		case stateLineIdentity:
			if field, err = buf.ReadBytes(' '); err == nil {
				if field = field[:len(field)-1]; !bytes.Equal(field, unavail) {
					e.Identity = string(field)
				}
			}
		case stateLineUserID:
			if field, err = buf.ReadBytes(' '); err == nil {
				if field = field[:len(field)-1]; !bytes.Equal(field, unavail) {
					e.UserID = string(field)
				}
			}
		case stateLineTimeBegin:
			_, err = buf.ReadBytes('[')
		case stateLineTime:
			if field, err = buf.ReadBytes(']'); err == nil {
				field = field[:len(field)-1]
				e.Time, err = time.Parse(layout, string(field))
			}
		case stateLineRequestBegin:
			_, err = buf.ReadBytes('"')
		case stateLineRequest:
			if field, err = buf.ReadBytes('"'); err == nil {
				field = field[:len(field)-1]
				e.Request, err = ParseRequest(field)
			}
		case stateLineRequestEnd:
			_, err = buf.ReadBytes(' ')
		case stateLineStatusCode:
			if field, err = buf.ReadBytes(' '); err == nil {
				field = field[:len(field)-1]
				var val int64
				if val, err = strconv.ParseInt(string(field), 10, 0); err == nil {
					e.StatusCode = int(val)
				}
			}
		case stateLineSize:
			if val := buf.String(); val != "-" {
				var i int64
				if i, err = strconv.ParseInt(buf.String(), 10, 0); err == nil {
					e.Size = int(i)
				}
			}
		}

		state++
	}

	return
}
