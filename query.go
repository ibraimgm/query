/*
Copyright (c) 2021, Rafael Ibraim Garcia Marques <ibraim.gm@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1.	Redistributions of source code must retain the above copyright notice, this
		list of conditions and the following disclaimer.

2.	Redistributions in binary form must reproduce the above copyright notice,
		this list of conditions and the following disclaimer in the documentation
		and/or other materials provided with the distribution.

3.	Neither the name of the copyright holder nor the names of its
		contributors may be used to endorse or promote products derived from
		this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

//Package query provides a bare bones, no-magic query builder.
package query

import (
	"reflect"
	"strconv"
	"strings"
)

// Builder is an auxiliary buffer of strings to make it a little
// easier to build dynamic queries from scratch.
type Builder struct {
	selectSQL strings.Builder
	fromSQL   strings.Builder
	whereSQL  strings.Builder
	orderSQL  strings.Builder

	params []interface{}
}

// Params returns the current slice of builder parameters.
func (b *Builder) Params() []interface{} {
	return b.params
}

// Add unconditionally append a sql string into this builder's buffer.
// If any values are supplyed, the parameters marked as '?' will be
// changed for the appropriate token.
func (b *Builder) Add(sql string, values ...interface{}) {
	s := b.loadParameters(sql, values)
	b.appendSQL(&b.selectSQL, s)
}

// AddIf appends an parametrized sql string only if the provided value is
// non-nil.
func (b *Builder) AddIf(sql string, value interface{}) {
	b.appendSQL(&b.selectSQL, b.addParam(sql, value))
}

// From appends the sql string in a special buffer, used to mount the
// 'from' clause. Otherwise, it has the same mechanics as the Add method.
func (b *Builder) From(sql string, values ...interface{}) {
	s := b.loadParameters(sql, values)
	b.appendSQL(&b.fromSQL, s)
}

// Order unconditionally adds the string to the 'order by' buffer.
func (b *Builder) Order(sql string) {
	b.appendSQL(&b.orderSQL, sql)
}

// Where appends the sql string in a special buffer, used to mount the
// 'where' clause. Otherwise, it has the same mechanics as the Add method.
func (b *Builder) Where(sql string, values ...interface{}) {
	s := b.loadParameters(sql, values)
	b.appendSQL(&b.whereSQL, s)
}

// WhereIf has the same functionality of AddIf, but writes to the special
// 'where' buffer.
func (b *Builder) WhereIf(sql string, value interface{}) {
	b.appendSQL(&b.whereSQL, b.addParam(sql, value))
}

// SetParam set the value of a positional parameter.
func (b *Builder) SetParam(index int, value interface{}) {
	for {
		if index <= len(b.params) {
			break
		}

		b.params = append(b.params, nil)
	}

	b.params[index-1] = value
}

// String implements the Stringer interface.
// This method returns the sql command to be executed.
func (b *Builder) String() string {
	return b.buildSQL()
}

func (b *Builder) loadParameters(originalSQL string, values []interface{}) string {
	newSQL := originalSQL
	psize := len(b.params)

	for _, value := range values {
		b.params = append(b.params, value)
		psize++
		newSQL = strings.Replace(newSQL, "?", "$"+strconv.Itoa(psize), 1)
	}

	return newSQL
}

func (b *Builder) addParam(sql string, value interface{}) string {
	if value == nil {
		return ""
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Ptr && v.IsNil() {
		return ""
	}

	b.params = append(b.params, value)
	return strings.Replace(sql, "?", "$"+strconv.Itoa(len(b.params)), 1)
}

func (b *Builder) appendSQL(sb *strings.Builder, s string) {
	if s == "" {
		return
	}

	sb.WriteString(s)
	// sb.WriteString(" ") // Maybe do this automatically?
}

func (b *Builder) buildSQL() string {
	return b.selectSQL.String() + b.fromSQL.String() + b.whereSQL.String() + b.orderSQL.String()
}
