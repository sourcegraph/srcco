.PHONY: install testserve serve 

install:
	go install github.com/samertm/srcco

testserve: install
	rm -rf .srclib-cache/ #SAMER: fix this
	src-docco -v gen
	cd site && python2 -m SimpleHTTPServer

serve: install
	cd site && python2 -m SimpleHTTPServer
