.PHONY: install testserve serve 

install:
	go-bindata ./data/
	go install sourcegraph.com/sourcegraph/srcco

testserve: install
	rm -rf .srclib-cache/ #SAMER: fix this
	srcco -v .
	cd docs && python2 -m SimpleHTTPServer

serve: install
	cd docs && python2 -m SimpleHTTPServer
