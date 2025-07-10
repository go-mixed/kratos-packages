package stream

import (
	"bufio"
	"encoding/json"
	"errors"
	protojson "github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	SseContentType = "text/event-stream"
)

type Sse struct {
	Id      string
	Event   string
	Data    string
	Retry   time.Duration
	Comment string
	Extra   map[string]string
}

type sseBuilder struct {
	id      string
	event   string
	data    string
	retry   time.Duration
	comment string
	extra   map[string]string
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

func (s *sseBuilder) Extra(key, value string) *sseBuilder {
	if s.extra == nil {
		s.extra = make(map[string]string)
	}
	s.extra[key] = value
	return s

}

func (s *sseBuilder) Build() Sse {
	return Sse{
		Id:      s.id,
		Event:   s.event,
		Data:    s.data,
		Retry:   s.retry,
		Comment: s.comment,
		Extra:   s.extra,
	}
}

func (s *Sse) Reset() {
	s.Id = ""
	s.Event = ""
	s.Data = ""
	s.Retry = 0
	s.Comment = ""
	s.Extra = nil
}

func (s *Sse) HasExtraKey(key string) bool {
	if s.Extra == nil {
		return false
	}
	_, ok := s.Extra[key]
	return ok
}

func (s *Sse) ExtraValue(key string) string {
	if s.Extra == nil {
		return ""
	}
	return s.Extra[key]
}

func (s *Sse) IsEmpty() bool {
	return s.Id == "" && s.Event == "" && s.Data == "" && s.Retry == 0 && s.Comment == "" && len(s.Extra) == 0
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

	for k, v := range s.Extra {
		result += k + ": " + v + "\n"
	}

	return result + "\n"
}

var separatorRegex = regexp.MustCompile(": ?")
var ErrSseContentType = errors.New("unexpected content type")
var ErrSseStatus = errors.New("unexpected status code")

// SSEReader reads Server-Sent Events from an HTTP response and calls the callback for each event.
func SSEReader(response *http.Response, callback func(sse Sse) error) error {
	if response.StatusCode != 200 {
		return ErrSseStatus
	} else if contentType := response.Header.Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		return ErrSseContentType
	}

	buf := bufio.NewReader(response.Body)
	var sse Sse
	for {
		line, err := buf.ReadString('\n')
		if line == "\n" || line == "\r\n" { // double \n means end of event and data
			if !sse.IsEmpty() {
				if err = callback(sse); err != nil {
					return err
				}
				sse.Reset()
			}
		} else if len(line) > 0 {
			// 即使在 Windows 等默认使用 \r\n（CRLF）作为换行符的系统中，SSE 传输时也要求统一使用 \n 作为行终止符，多余的 \r 会被视为字段值的一部分（而非换行符）。
			// remove trailing \n
			line = line[:len(line)-1]
		}
		if err != nil && err != io.EOF {
			return err
		}
		segments := separatorRegex.Split(line, 2)
		if len(segments) == 2 {
			switch strings.ToLower(segments[0]) {
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
				if strings.TrimSpace(segments[0]) == "" { // line starts with a colon
					sse.Comment = segments[1]
				} else { // extra fields
					if sse.Extra == nil {
						sse.Extra = make(map[string]string)
					}
					sse.Extra[segments[0]] = segments[1]
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// LineReader 不规范的流输出，使用\n做的分隔符
func LineReader(response *http.Response, callback func(sse Sse) error) error {
	if response.StatusCode != 200 {
		return ErrSseStatus
	}

	buf := bufio.NewReader(response.Body)
	var sse Sse
	for {
		line, err := buf.ReadString('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if len(line) > 0 {
			line = line[:len(line)-1] // remove trailing \n
		}
		sse.Data = line
		if err = callback(sse); err != nil {
			return err
		}
	}
	return nil
}
