/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePlainText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "only whitespace", in: " \n\t\r\n ", want: ""},
		{name: "trim surrounding whitespace", in: "  remember to retry  ", want: "remember to retry"},
		{name: "windows line endings", in: "first\r\nsecond\rthird", want: "first\nsecond\nthird"},
		{name: "collapse blank lines", in: "a\n\n\n\n\nb", want: "a\n\nb"},
		{name: "blank line with spaces is collapsed", in: "a\n  \n \n\nb", want: "a\n\nb"},
		{name: "trailing spaces on lines", in: "a   \nb\t", want: "a\nb"},
		{name: "keep indentation and tabs", in: "list:\n\t- one\n  - two", want: "list:\n\t- one\n  - two"},
		{name: "drop control characters", in: "a\x00b\x1bc\u0085d", want: "abcd"},
		{name: "unicode is kept", in: "Çok güzel 👍 答案", want: "Çok güzel 👍 答案"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizePlainText(tt.in))
		})
	}
}
