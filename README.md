Introduction
============

Gypsy is a simplified YAML parser written in Go.  It is intended to be used as
a simple configuration file, and as such does not support a lot of the more
nuanced syntaxes allowed in full-fledged YAML.

Syntax
======

Lists
-----

Gypsy understands the following to be a list:
    - one
    - two
    - three

This is parsed as a `yaml.List`, and can be retrieved from the
`yaml.Node.List()` method.  In this case, each element of the `yaml.List` would
be a `yaml.Scalar` whose value can be retrieved with the `yaml.Scalar.String()`
method.

Mappings
--------

Gypsy understands the following to be a mapping:
    key:     value
    foo:     bar
    running: away

A mapping is an unordered list of `key:value` pairs.  All whitespace after the
colon is stripped from the value and is used for alignment purposes during
export.  If the value is not a list or a map, everything after the first
non-space character until the end of the line is used as the `yaml.Scalar`
value.

Combinations
------------

Gypsy allows arbitrary nesting of maps inside lists, lists inside of maps, and
maps and/or lists nested inside of themselves.

A map inside of a list:
    - name: John Smith
      age:  42
    - name: Jane Smith
      age:  45

A list inside of a map:
    schools:
      - Meadow Glen
      - Forest Creek
      - Shady Grove
    libraries:
      - Joseph Hollingsworth Memorial
      - Andrew Keriman Memorial

A list of lists:
    - - one
      - two
      - three
    - - un
      - deux
      - troix
    - - ichi
      - ni
      - san

A map of maps
    google:
      company: Google, Inc.
      ticker:  GOOG
      url:     http://google.com/
    yahoo:
      company: Yahoo, Inc.
      ticker:  YHOO
      url:     http://yahoo.com/
