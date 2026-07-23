/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

 package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestWrapNilReturnsNil(t *testing.T) {
	if Wrap(KindAPI, nil) != nil {
		t.Fatal("Wrap of a nil error should return nil")
	}
}

func TestKindOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"nil", nil, KindGeneric},
		{"plain error", stderrors.New("boom"), KindGeneric},
		{"wrapped", Wrap(KindConnection, stderrors.New("refused")), KindConnection},
		{"wrapped again", fmt.Errorf("outer: %w", Wrap(KindNotFound, stderrors.New("404"))), KindNotFound},
	}
	for _, c := range cases {
		if got := KindOf(c.err); got != c.want {
			t.Errorf("%s: KindOf = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestWrapfPreservesKindAndMessage(t *testing.T) {
	err := Wrapf(KindEnvironment, "docker unreachable on attempt %d", 2)
	if KindOf(err) != KindEnvironment {
		t.Errorf("KindOf = %v, want KindEnvironment", KindOf(err))
	}
	if got := err.Error(); got != "docker unreachable on attempt 2" {
		t.Errorf("Error() = %q", got)
	}
}
