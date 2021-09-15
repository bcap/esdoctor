package diagnosis

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type CommentType string

const Info CommentType = "info"
const Summary CommentType = "summary"
const Advice CommentType = "advice"
const Warning CommentType = "warning"

var allTypes map[CommentType]struct{} = map[CommentType]struct{}{
	Info:    {},
	Summary: {},
	Advice:  {},
	Warning: {},
}

type Comment struct {
	Time    time.Time   `json:"time"`
	Type    CommentType `json:"type"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
}

var codePattern = regexp.MustCompile(`^(([ISAW])\d{3}):`)

func NewComment(when *time.Time, msg string, args ...interface{}) Comment {
	code := "????"
	typ := CommentType("?")
	matches := codePattern.FindStringSubmatch(msg)
	if len(matches) == 3 {
		code = matches[1]
		switch matches[2] {
		case "I":
			typ = Info
		case "S":
			typ = Summary
		case "A":
			typ = Advice
		case "W":
			typ = Warning
		}
		msg = strings.TrimLeft(msg[len(matches[0]):], " \t")
	} else {
		log.Errorf(
			"Cannot infer comment level: template does not seem to respect the %q format: %q",
			codePattern, msg,
		)
	}
	if when == nil || when.IsZero() {
		now := time.Now()
		when = &now
	}
	return Comment{
		Type:    typ,
		Code:    code,
		Time:    *when,
		Message: fmt.Sprintf(msg, args...),
	}
}

func (d *Diagnostics) Comment(msg string, args ...interface{}) {
	d.AddComment(NewComment(nil, msg, args...))
}

func (d *Diagnostics) AddComment(c Comment) {
	d.commentLock.Lock()
	d.comments = append(d.comments, c)
	d.commentLock.Unlock()
	if d.config.writer != nil {
		if err := d.config.writer.Write(d, c); err != nil {
			log.Errorf("Failed to write comment %s: %v", c, err)
		}
	}
}

func (d *Diagnostics) Comments() []Comment {
	d.commentLock.RLock()
	result := make([]Comment, len(d.comments))
	copy(result, d.comments)
	d.commentLock.RUnlock()
	return result
}

//
// CommentWriter
//

type CommentWriter interface {
	Begin(*Diagnostics) error
	Write(*Diagnostics, Comment) error
	End(*Diagnostics) error
}

func NewTextCommentWriter(writer io.Writer, types []CommentType, coloured bool) CommentWriter {
	typesMap := allTypes
	if types != nil {
		typesMap = map[CommentType]struct{}{}
		for _, t := range types {
			typesMap[t] = struct{}{}
		}
	}
	return &textCommentWriter{
		writer:   writer,
		types:    typesMap,
		coloured: coloured,
	}
}

func NewJSONCommentWriter(writer io.Writer, dump bool) CommentWriter {
	return &jsonCommentWriter{
		writer: writer,
		dump:   dump,
	}
}

type textCommentWriter struct {
	writer    io.Writer
	types     map[CommentType]struct{}
	coloured  bool
	infos     int32
	summaries int32
	advices   int32
	warnings  int32
}

type colorFn = func(string, ...interface{}) string

var warningColor colorFn = color.RedString
var adviceColor colorFn = color.YellowString
var summaryColor colorFn = color.HiBlueString

func (t *textCommentWriter) Begin(*Diagnostics) error {
	return nil
}

func (t *textCommentWriter) Write(_ *Diagnostics, c Comment) error {
	var err error
	if _, ok := t.types[c.Type]; ok {
		codeString := c.Code
		if t.coloured {
			switch c.Type {
			case Warning:
				codeString = warningColor(codeString)
			case Advice:
				codeString = adviceColor(codeString)
			case Summary:
				codeString = summaryColor(codeString)
			}
		}
		_, err = fmt.Fprintf(t.writer, "%s %s\n", codeString, c.Message)
	}
	switch c.Type {
	case Info:
		atomic.AddInt32(&t.infos, 1)
	case Summary:
		atomic.AddInt32(&t.summaries, 1)
	case Advice:
		atomic.AddInt32(&t.advices, 1)
	case Warning:
		atomic.AddInt32(&t.warnings, 1)
	}
	return err
}

func (t *textCommentWriter) End(*Diagnostics) error {
	infos := atomic.LoadInt32(&t.infos)
	summaries := atomic.LoadInt32(&t.summaries)
	advices := atomic.LoadInt32(&t.advices)
	warnings := atomic.LoadInt32(&t.warnings)
	resultStr := "Result"
	if t.coloured {
		if warnings > 0 {
			resultStr = warningColor(resultStr)
		} else if advices > 0 {
			resultStr = adviceColor(resultStr)
		} else if summaries > 0 {
			resultStr = summaryColor(resultStr)
		}
	}
	_, err := fmt.Fprintf(
		t.writer,
		"%s: %d warnings, %d advices, %d summaries and %d informational comments\n",
		resultStr, warnings, advices, summaries, infos,
	)
	return err
}

type jsonCommentWriter struct {
	writer io.Writer
	dump   bool
}

func (*jsonCommentWriter) Begin(*Diagnostics) error {
	return nil
}

func (*jsonCommentWriter) Write(_ *Diagnostics, c Comment) error {
	return nil
}

func (j *jsonCommentWriter) End(d *Diagnostics) error {
	if j.dump {
		return d.JSONDump(j.writer)
	} else {
		encoder := json.NewEncoder(j.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(d.Comments())
	}
}
