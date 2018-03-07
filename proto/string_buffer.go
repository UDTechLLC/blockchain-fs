package main

import (
	"fmt"
	"strings"
)

const maxLength = 200

/// a java style string buffer
type StringBuffer struct {
	data   []string
	index  int
	indent int
}

func NewStringBuffer() *StringBuffer {
	ret := new(StringBuffer)
	ret.index = 0
	ret.indent = 0
	ret.data = make([]string, maxLength)
	return ret
}

/// append a string to the tail of this buffer
func (sb *StringBuffer) Append(str string) *StringBuffer {
	sb.data[sb.index] = strings.Repeat("\t", sb.indent) + str
	sb.index++
	if sb.index >= maxLength {
		sb.slice()
	}
	return sb
}

func (sb *StringBuffer) AppendFormat(str string, args ...interface{}) *StringBuffer {
	return sb.Append(fmt.Sprintf(str, args...))
}

func (sb *StringBuffer) AppendLine(str string) *StringBuffer {
	return sb.Append(str + "\n")
}

func (sb *StringBuffer) AppendLineIndent(str string) *StringBuffer {
	sb.AppendLine(str)
	sb.indent++
	return sb
}

func (sb *StringBuffer) AppendLineClose(str string) *StringBuffer {
	sb.indent--
	return sb.AppendLine(str)
}

func (sb *StringBuffer) AppendClose(str string) *StringBuffer {
	sb.indent--
	return sb.Append(str)
}

func (sb *StringBuffer) slice() *StringBuffer {
	sb.data[0] = sb.ToString()
	sb.index = 1
	return sb
}

/// append a line as comment.
func (sb *StringBuffer) AppendComment(str string) *StringBuffer {
	return sb.AppendLine("/// " + str)
}

func (sb *StringBuffer) ToString() string {
	return strings.Join(sb.data, "")
}

/// clear elements.
func (sb *StringBuffer) Clear() {
	sb.indent = 0
}
