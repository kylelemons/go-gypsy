include $(GOROOT)/src/Make.inc

TARG=goyaml
GOFILES=\
	types.go\
	yaml.go\

include $(GOROOT)/src/Make.pkg
