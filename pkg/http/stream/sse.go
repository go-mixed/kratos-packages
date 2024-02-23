package stream

import (
	"bufio"
	"encoding/json"
	"errors"
	protojson "github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Sse struct {
	Id      string
	Event   string
	Data    string
	Retry   time.Duration
	Comment string
}

type sseBuilder struct {
	id      string
	event   string
	data    string
	retry   time.Duration
	comment string
}

func NewSseBuilder() *sseBuilder {
	return &sseBuilder{}
}

func (s *sseBuilder) Id(id string) *sseBuilder {
	s.id = id
	return s
}

func (s *sseBuilder) Event(event string) *sseBuilder {
	s.event = event
	return s
}

func (s *sseBuilder) Data(data string) *sseBuilder {
	s.data += data
	return s
}

func (s *sseBuilder) Json(data any) *sseBuilder {
	j, _ := json.Marshal(data)
	s.data += string(j)
	return s
}

func (s *sseBuilder) Protobuf(data proto.Message) *sseBuilder {
	j, _ := protojson.MarshalOptions.Marshal(data)
	s.data += string(j)
	return s
}

func (s *sseBuilder) Retry(retry time.Duration) *sseBuilder {
	s.retry = retry
	return s
}

func (s *sseBuilder) Comment(comment string) *sseBuilder {
	s.comment = comment
	return s
}

func (s *sseBuilder) Build() Sse {
	return Sse{
		Id:      s.id,
		Event:   s.event,
		Data:    s.data,
		Retry:   s.retry,
		Comment: s.comment,
	}
}

func (s *Sse) Reset() {
	s.Id = ""
	s.Event = ""
	s.Data = ""
	s.Retry = 0
	s.Comment = ""
}

func (s *Sse) IsEmpty() bool {
	return s.Id == "" && s.Event == "" && s.Data == "" && s.Retry == 0 && s.Comment == ""
}

func (s *Sse) String() string {
	var result string
	if s.Id != "" {
		result += "id: " + s.Id + "\n"
	}
	if s.Event != "" {
		result += "event: " + s.Event + "\n"
	}
	if s.Retry > 0 {
		result += "retry: " + strconv.FormatInt(s.Retry.Milliseconds(), 10) + "\n"
	}
	if s.Comment != "" {
		result += ":" + s.Comment + "\n"
	}
	if s.Data != "" {
		result += "data: " + s.Data + "\n"
	}
	return result + "\n"
}

// SSEReader reads Server-Sent Events from an HTTP response and calls the callback for each event.
func SSEReader(response *http.Response, callback func(sse Sse) error) error {
	if response.StatusCode != 200 {
		return errors.New("invalid status code: " + response.Status)
	} else if contentType := response.Header.Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		return errors.New("unexpected content type: " + contentType)
	}

	buf := bufio.NewReader(response.Body)
	var sse Sse
	for {
		line, err := buf.ReadString('\n')
		if line == "\n" { // double \n means end of event and data
			if !sse.IsEmpty() {
				if err = callback(sse); err != nil {
					return err
				}
				sse.Reset()
			}
		} else if len(line) > 0 {
			line = line[:len(line)-1] // remove trailing \n
		}
		if err != nil && err != io.EOF {
			return err
		}
		segments := strings.SplitN(line, ": ", 2)
		if len(segments) == 2 {
			switch segments[0] {
			case "id":
				sse.Id = segments[1]
			case "retry":
				_retry, _ := strconv.Atoi(segments[1])
				sse.Retry = time.Duration(_retry) * time.Millisecond
			case "event":
				sse.Event = segments[1]
			case "data":
				sse.Data += segments[1] // append to data
			default: // unknown field or comment
				if segments[0] == "" { // line starts with a colon
					sse.Comment = segments[1]
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}
