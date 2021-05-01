package ami

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	amimodels "github.com/wenerme/astgo/pkg/ami/models"
	"io"
	"strings"
)

type MessageType string

const (
	MessageTypeAction   MessageType = "Action"
	MessageTypeEvent    MessageType = "Event"
	MessageTypeResponse MessageType = "Response"
)

type Message struct {
	Type       MessageType // type of message
	Name       string      // name of message
	Attributes map[string]interface{}
}

func (m *Message) Read(r *bufio.Reader) (err error) {
	m.Attributes = map[string]interface{}{}
	var line string
	line, err = r.ReadString('\n')
	if err != nil {
		return err
	}
	line = strings.TrimSuffix(line, "\r\n")
	sp := strings.SplitN(line, ":", 2)
	if len(sp) != 2 {
		return errors.Errorf("invalid line read: %q", line)
	}
	m.Type = MessageType(sp[0])
	m.Name = strings.TrimSpace(sp[1])
	switch m.Type {
	case MessageTypeAction:
	case MessageTypeEvent:
	case MessageTypeResponse:
	default:
		return errors.Errorf("invalid message type: %q", sp[0])
	}

	for {
		line, err = r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSuffix(line, "\r\n")
		if len(line) == 0 {
			break
		}
		sp = strings.SplitN(line, ":", 2)
		if len(sp) != 2 {
			return errors.Errorf("invalid line read: %q", line)
		}
		m.Attributes[sp[0]] = strings.TrimSpace(sp[1])
	}

	return
}
func (m *Message) Write(w io.Writer) (err error) {
	wr := bufio.NewWriter(w)
	_, _ = wr.WriteString(fmt.Sprintf("%v: %v\r\n", m.Type, m.Name))
	for k, v := range m.Attributes {
		_, _ = wr.WriteString(fmt.Sprintf("%v: %v\r\n", k, v))
	}
	_, _ = wr.WriteString("\r\n")
	err = wr.Flush()
	return
}

func (m Message) Format() string {
	b := bytes.Buffer{}
	_ = m.Write(&b)
	return b.String()
}

func (m *Message) AttrString(name string) string {
	if m.Attributes == nil {
		return ""
	}
	msg := m.Attributes[name]
	if msg == nil {
		return ""
	}
	return fmt.Sprint(msg)
}
func (m *Message) Message() string {
	return m.AttrString("Message")
}

func (m *Message) Success() bool {
	if m.Type == MessageTypeResponse && m.Name == "Success" {
		return true
	}
	return false
}

func (m *Message) Error() error {
	if m.Type == MessageTypeResponse && m.Name == "Error" {
		msg := m.Message()
		if msg == "" {
			msg = "error response"
		}
		return errors.New(msg)
	}
	return nil
}
func (m *Message) SetAttr(name string, val interface{}) {
	if m.Attributes == nil {
		m.Attributes = make(map[string]interface{})
	}
	m.Attributes[name] = val
}

func MustConvertToMessage(in interface{}) (msg *Message) {
	m, err := ConvertToMessage(in)
	if err != nil {
		panic(err)
	}
	return m
}
func ConvertToMessage(in interface{}) (msg *Message, err error) {
	msg = &Message{}
	switch a := in.(type) {
	case Message:
		return &a, err
	case *Message:
		return a, err
	case amimodels.Action:
		msg.Type = MessageTypeAction
		msg.Name = a.ActionTypeName()
	case amimodels.Event:
		msg.Type = MessageTypeEvent
		msg.Name = a.EventTypeName()
	default:
		return nil, errors.Errorf("invalid type: %T", in)
	}
	m := make(map[string]interface{})
	err = mapstructure.Decode(in, &m)
	// remove zero
	for k, v := range m {
		rm := v == nil
		switch v := v.(type) {
		case string:
			rm = v == ""
		}
		if rm {
			delete(m, k)
		}
	}
	msg.Attributes = m
	return
}
