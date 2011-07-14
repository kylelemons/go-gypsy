// Copyright (C) 2011, Kyle Lemons
// All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 
// * Redistributions of source code must retain the above copyright notice,
//   this list of conditions and the following disclaimer.
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
// * Neither the name of the author nor the names of its contributors may be
//   used to endorse or promote products derived from this software without
//   specific prior written permission.
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
//
// Contributors:
//   andybalholm

// Gypsy is a simplified YAML parser written in Go.  It is intended to be used as
// a simple configuration file, and as such does not support a lot of the more
// nuanced syntaxes allowed in full-fledged YAML.  YAML does not allow indent with
// tabs, and GYPSY does not ever consider a tab to be a space character.  It is
// recommended that your editor be configured to convert tabs to spaces when
// editing Gypsy config files.
// 
// Gypsy understands the following to be a list:
// 
//     - one
//     - two
//     - three
// 
// This is parsed as a `yaml.List`, and can be retrieved from the
// `yaml.Node.List()` method.  In this case, each element of the `yaml.List` would
// be a `yaml.Scalar` whose value can be retrieved with the `yaml.Scalar.String()`
// method.
// 
// Gypsy understands the following to be a mapping:
// 
//     key:     value
//     foo:     bar
//     running: away
// 
// A mapping is an unordered list of `key:value` pairs.  All whitespace after the
// colon is stripped from the value and is used for alignment purposes during
// export.  If the value is not a list or a map, everything after the first
// non-space character until the end of the line is used as the `yaml.Scalar`
// value.
// 
// Gypsy allows arbitrary nesting of maps inside lists, lists inside of maps, and
// maps and/or lists nested inside of themselves.
// 
// A map inside of a list:
// 
//     - name: John Smith
//       age:  42
//     - name: Jane Smith
//       age:  45
// 
// A list inside of a map:
// 
//     schools:
//       - Meadow Glen
//       - Forest Creek
//       - Shady Grove
//     libraries:
//       - Joseph Hollingsworth Memorial
//       - Andrew Keriman Memorial
// 
// A list of lists:
// 
//     - - one
//       - two
//       - three
//     - - un
//       - deux
//       - trois
//     - - ichi
//       - ni
//       - san
// 
// A map of maps:
// 
//     google:
//       company: Google, Inc.
//       ticker:  GOOG
//       url:     http://google.com/
//     yahoo:
//       company: Yahoo, Inc.
//       ticker:  YHOO
//       url:     http://yahoo.com/
// 
// In the case of a map of maps, all sub-keys must be on subsequent lines and
// indented equally.  It is allowable for the first key/value to be on the same
// line if there is more than one key/value pair, but this is not recommended.
// 
// Values can also be expressed in long form (leading whitespace of the first line
// is removed from it and all subsequent lines).  In the normal (baz) case,
// newlines are treated as spaces, all indentation is removed.  In the folded case
// (bar), newlines are treated as spaces, except pairs of newlines (e.g. a blank
// line) are treated as a single newline, only the indentation level of the first
// line is removed, and newlines at the end of indented lines are preserved.  In
// the verbatim (foo) case, only the indent at the level of the first line is
// stripped.  The example:
// 
//     foo: |
//       lorem ipsum dolor
//       sit amet
//     bar: >
//       lorem ipsum
// 
//         dolor
// 
//       sit amet
//     baz:
//       lorem ipsum
//        dolor sit amet
// 
// The YAML subset understood by Gypsy can be expressed (loosely) in the following
// grammar (not including comments):
// 
//               OBJECT = MAPPING | SEQUENCE | SCALAR .
//         SHORT-OBJECT = SHORT-MAPPING | SHORT-SEQUENCE | SHORT-SCALAR .
//                  EOL = '\n'
// 
//              MAPPING = { LONG-MAPPING | SHORT-MAPPING } .
//             SEQUENCE = { LONG-SEQUENCE | SHORT-SEQUENCE } .
//               SCALAR = { LONG-SCALAR | SHORT-SCALAR } .
// 
//         LONG-MAPPING = { INDENT KEY ':' OBJECT EOL } .
//        SHORT-MAPPING = '{' KEY ':' SHORT-OBJECT { ',' KEY ':' SHORT-OBJECT } '}' EOL .
// 
//        LONG-SEQUENCE = { INDENT '-' OBJECT EOL } EOL .
//       SHORT-SEQUENCE = '[' SHORT-OBJECT { ',' SHORT-OBJECT } ']' EOL .
//       
//          LONG-SCALAR = ( '|' | '>' | ) EOL { INDENT SHORT-SCALAR EOL }
//         SHORT-SCALAR = { alpha | digit | punct | ' ' | '\t' } .
// 
//                  KEY = { alpha | digit }
//               INDENT = { ' ' }
//
// Any line where the first non-space character is a sharp sign (#) is a comment.
// It will be ignored.
// Only full-line comments are allowed.
package yaml

// BUG(kevlar): Multi-line strings are currently not supported.
