// Copyright 2013 Google, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml

import (
	"testing"
)

var dummyConfigFile = `
mapping:
  key1: value1
  key2: value2
  key3: 5
  key4: true
  key5: false
list:
  - item1
  - item2
config:
  server:
    - www.google.com
    - www.cnn.com
    - www.example.com
  admin:
    - username: god
      password: z3u5
    - username: lowly
      password: f!r3m3
`

var configGetTests = []struct {
	Spec        string
	Want        string
	Err         string
	FormatSpec  string
	FormatParam int
}{
	{"mapping.key1", "value1", "", "mapping.key%d", 1},
	{"mapping.key2", "value2", "", "mapping.key%d", 2},
	{"list[0]", "item1", "", "list[%d]", 0},
	{"list[1]", "item2", "", "list[%d]", 1},
	{"list", "", `yaml: list: type mismatch: "list" is yaml.List, want yaml.Scalar (at "$")`, "", 0},
	{"list.0", "", `yaml: .list.0: type mismatch: ".list" is yaml.List, want yaml.Map (at ".0")`, "list.%d", 0},
	{"config.server[0]", "www.google.com", "", "config.server[%d]", 0},
	{"config.server[1]", "www.cnn.com", "", "config.server[%d]", 1},
	{"config.server[2]", "www.example.com", "", "config.server[%d]", 2},
	{"config.server[3]", "", `yaml: .config.server[3]: ".config.server[3]" not found`, "config.server[%d]", 3},
	{"config.listen[0]", "", `yaml: .config.listen[0]: ".config.listen" not found`, "config.listen[%d]", 0},
	{"config.admin[0].username", "god", "", "config.admin[%d].username", 0},
	{"config.admin[1].username", "lowly", "", "config.admin[%d].username", 1},
	{"config.admin[2].username", "", `yaml: .config.admin[2].username: ".config.admin[2]" not found`, "config.admin[%d].username", 2},
}

func TestGet(t *testing.T) {
	config := Config(dummyConfigFile)

	for _, test := range configGetTests {
		got, err := config.Get(test.Spec)
		if want := test.Want; got != want {
			t.Errorf("Get(%q) = %q, want %q", test.Spec, got, want)
		}

		switch err {
		case nil:
			got = ""
		default:
			got = err.Error()
		}
		if want := test.Err; got != want {
			t.Errorf("Get(%q) error %#q, want %#q", test.Spec, got, want)
		}

		if test.FormatSpec != "" {
			got, err = config.Get(test.FormatSpec, test.FormatParam)
			if want := test.Want; got != want {
				t.Errorf("Get(%q) = %q, want %q", test.Spec, got, want)
			}

			switch err {
			case nil:
				got = ""
			default:
				got = err.Error()
			}
			if want := test.Err; got != want {
				t.Errorf("Get(%q) error %#q, want %#q", test.Spec, got, want)
			}
		}
	}

	i, err := config.GetInt("mapping.key3")
	if err != nil || i != 5 {
		t.Errorf("GetInt mapping.key3 wrong")
	}
	i, err = config.GetInt("mapping.key%d", 3)
	if err != nil || i != 5 {
		t.Errorf("GetInt mapping.key3 wrong")
	}

	b, err := config.GetBool("mapping.key4")
	if err != nil || b != true {
		t.Errorf("GetBool mapping.key4 wrong")
	}
	b, err = config.GetBool("mapping.key%d", 4)
	if err != nil || b != true {
		t.Errorf("GetBool mapping.key4 wrong")
	}

	b, err = config.GetBool("mapping.key5")
	if err != nil || b != false {
		t.Errorf("GetBool mapping.key5 wrong")
	}
	b, err = config.GetBool("mapping.key%d", 5)
	if err != nil || b != false {
		t.Errorf("GetBool mapping.key5 wrong")
	}

}
