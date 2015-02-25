package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path"
	"path/filepath"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindata_file_info struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindata_file_info) Name() string {
	return fi.name
}
func (fi bindata_file_info) Size() int64 {
	return fi.size
}
func (fi bindata_file_info) Mode() os.FileMode {
	return fi.mode
}
func (fi bindata_file_info) ModTime() time.Time {
	return fi.modTime
}
func (fi bindata_file_info) IsDir() bool {
	return false
}
func (fi bindata_file_info) Sys() interface{} {
	return nil
}

var _data_publish_gh_pages_sh = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x8c\x57\x6d\x73\xda\x48\x12\xfe\xae\x5f\xd1\xc1\x54\x6c\x27\xbc\xc4\xf9\x78\x59\x92\x53\xb0\x6c\xab\x0e\x83\x4b\xe0\xf5\xa5\xd6\x7b\x41\x48\x23\x4b\x17\xa1\xd1\x6a\x46\xd8\xdc\x26\xff\xfd\x9e\x9e\x11\x18\x58\xd7\x5d\x5c\x2e\x40\x33\xfd\xfe\xf2\x74\xeb\xe8\x55\xbf\x56\x55\x7f\x91\x15\x7d\x51\xac\x68\x11\xaa\xd4\x71\x8e\x68\x28\xcb\x75\x95\x3d\xa4\x9a\x4e\xa2\x53\x7a\xff\xee\xec\x8c\xfc\xb0\xa0\xeb\x30\x1a\x09\x19\xd3\x2f\x59\x58\xfc\xbd\x10\xab\xac\xea\x15\x42\x7f\x74\x8e\xc0\x72\x23\xaa\x65\xa6\x54\x26\x0b\xca\x14\xa5\xa2\x12\x8b\x35\x3d\x54\x61\xa1\x45\xdc\xa1\xa4\x12\x82\x64\x42\x51\x1a\x56\x0f\xa2\x43\x5a\x52\x58\xac\xa9\x14\x95\x02\x83\x5c\xe8\x30\x2b\xb2\xe2\x81\x42\x8a\xa0\x9a\x29\x75\x0a\x31\x4a\x26\xfa\x31\xac\x04\x88\x63\xe8\x08\x95\x92\x51\x16\x42\x22\xc5\x32\xaa\x97\xa2\xd0\xa1\x66\x8d\x49\x96\x0b\x45\x27\x3a\x15\xd4\x9a\x36\x3c\xad\x53\xa3\x26\x16\x61\x4e\x59\x41\x7c\xb7\xb9\xa2\xc7\x4c\xa7\xb2\xd6\x54\x09\xa5\xab\x2c\x62\x19\x1d\xc8\xcf\x8a\x28\xaf\x63\xb6\x63\x43\x90\x67\xcb\xac\xd1\xc1\x02\x4c\x4c\x14\x8b\xad\x15\xbc\x60\x5b\x3b\xb4\x94\x71\x96\xf0\xb7\x30\xae\x95\xf5\x22\xcf\x54\xda\xa1\x38\x63\xe1\x8b\x5a\x0b\x96\xad\xf8\x38\x12\x05\xf3\xc1\x9b\xbe\xac\x48\x89\x3c\x67\x19\x19\x6c\x37\x1e\x3f\x5b\x68\x68\x58\x4f\xc9\x61\xd5\x4d\xa0\x8c\xe6\xc7\x54\x2e\xf7\xbd\xc9\x14\xe4\x27\x75\x55\x40\xad\x30\x5c\xb1\x44\xe8\x3a\xac\xf3\xdf\x22\xd2\x7c\xc2\x0c\x89\xcc\x73\xf9\xc8\xee\x45\xb2\x88\x33\xf6\x4a\xfd\xcd\x24\x6f\x86\xdb\x70\x21\x57\xc2\xb8\x64\x33\x5f\x48\x0d\x7b\xad\x1d\x9c\x8b\xf2\x39\xc1\xcd\x95\x4a\x43\x38\xb0\x10\x4d\xdc\xa0\x1a\x71\x0e\x77\x7c\xaa\xac\xdf\x4a\xa3\x0a\x32\xa4\xa1\x94\x95\x51\x7a\xe8\x6d\xcf\x1a\x71\xe5\xd1\x74\x72\x31\xbb\x73\x03\x8f\xfc\x29\xdd\x04\x93\x5f\xfd\x73\xef\x9c\x5a\xee\x14\xcf\xad\x0e\xdd\xf9\xb3\xab\xc9\xed\x8c\x40\x11\xb8\xe3\xd9\x17\x9a\x5c\x90\x3b\xfe\x42\xff\xf0\xc7\xe7\x1d\xf2\xfe\x79\x13\x78\xd3\x29\x4d\x02\xf2\xaf\x6f\x46\xbe\x87\x33\x7f\x3c\x1c\xdd\x9e\xfb\xe3\x4b\xfa\x7c\x3b\x83\x8e\xf1\x64\x46\x23\xff\xda\x9f\x41\xec\x6c\x62\x54\x36\xc2\x7c\x6f\xca\xe2\xae\xbd\x60\x78\x85\x47\xf7\xb3\x3f\xf2\x67\x5f\x3a\x74\xe1\xcf\xc6\x2c\xf5\x02\x62\x5d\xba\x71\x83\x99\x3f\xbc\x1d\xb9\x01\xdd\xdc\x06\x37\x93\xa9\x07\x03\xce\x8d\xe0\xb1\x3f\xbe\x08\xa0\xc9\xbb\xf6\xc6\xb3\x1e\x34\xe3\x8c\xbc\x5f\xf1\x40\xd3\x2b\x77\x34\x32\xca\xdc\x5b\x78\x10\x18\x1b\x87\x93\x9b\x2f\x81\x7f\x79\x35\xa3\xab\xc9\xe8\xdc\xc3\xe1\x67\x0f\xb6\xb9\x9f\x47\x9e\x55\x06\xc7\x86\x23\xd7\xbf\xe6\xea\x39\x77\xaf\xdd\x4b\xcf\xf0\x4d\x20\x27\x30\x84\x8d\x85\x77\x57\x9e\x39\x82\x46\x17\xff\xc3\x99\x3f\x19\xb3\x2b\xc3\xc9\x78\x16\xe0\xb1\x03\x4f\x83\xd9\x96\xf5\xce\x9f\x7a\x1d\x72\x03\x7f\xca\x61\xb9\x08\x26\xd7\x1d\x9a\x98\xe0\x80\x67\x62\xc4\x80\x73\xec\x59\x39\x1c\xf2\xfd\xcc\x80\x84\x9f\x6f\xa7\xde\xb3\x35\xe7\x9e\x3b\x82\xb4\x29\x33\xef\x12\xf7\x18\x4c\x46\x72\x85\x9a\xcb\xd7\xa4\xc3\x6f\x02\xbd\x5a\xa1\x7e\x1f\xd0\x5f\xf5\xa2\x17\xc9\x65\xdf\xe0\x48\xff\xa1\x92\x11\xa8\x95\xd0\xd4\x15\x74\x44\x53\x2d\x4b\x6a\xba\x2e\xc9\x2a\xa5\x29\x09\xb3\xbc\x46\xb5\xeb\x34\xd4\x24\xa3\xa8\xae\x94\xe3\x9c\x4f\x86\xd3\xaf\x37\xee\xec\x6a\xd0\x83\xc8\xbe\xaa\xa2\x48\x76\xf5\xb2\x74\x66\x6e\x70\xe9\xcd\xbe\x7e\x46\x6a\x87\x57\x83\x87\xb4\x5b\x86\x0f\x42\x39\xbf\xfd\x46\xed\x33\xfa\xfd\x77\x7a\xfd\x9a\x1a\x92\xc0\xbb\x9e\xcc\xbc\x01\x8e\xbf\x7f\x3f\x38\x93\xe8\x84\xac\x60\x1f\x2e\xd1\x83\xaa\xcc\xd0\xfb\x8c\x0a\x28\x68\x5d\x2b\x54\x7b\x22\xab\xa5\x45\x07\xfc\xb7\x95\x8e\x45\x55\xd9\xd6\x7d\x14\x68\xc2\xe2\x58\xd3\x23\x6a\x9f\x1b\xb0\x12\x79\xb8\xb6\xc6\x87\x0a\x24\x04\x52\x60\x80\x6d\x4d\x68\x00\xa0\x54\x3d\x42\x4b\x30\xeb\x63\x15\x96\x1c\x23\x23\x0a\xcd\x6c\x69\x53\x3c\xe5\xdc\xbe\x12\xbe\x8b\x7c\x25\x54\xaf\xd7\x73\xc4\x93\x88\xbe\x82\xf6\xe4\x94\xfe\x74\x88\x80\xaf\x6a\x70\x7c\x8c\x5f\x30\x8e\x4e\x4e\x28\xa3\x01\x9d\x7d\xc0\xd7\x2f\x03\x6a\x1f\xe1\xc7\xdb\xb7\x74\x7a\xfa\x01\x62\x41\x43\x24\x56\x68\x4c\x30\x0d\xee\xdb\xed\xcc\x9c\x64\x09\x71\x98\x70\x46\x83\x01\xbd\xb9\xa7\x37\x08\xd8\x07\x36\xb3\x30\xf7\x84\xf4\xfc\xa0\x3b\xc1\x70\x0c\x0c\xb6\x3e\xc5\x12\xf8\x26\xe8\x8f\x5a\x6a\x74\xfe\x63\x06\x14\x00\x4a\x30\x38\xc0\xb1\x98\x5d\x2e\xc3\x4a\x73\xe3\x43\xae\x41\x6e\x82\xe2\x1a\x2e\x34\x32\x8d\xdd\x2d\xd6\xaa\xe8\xde\x7c\xdf\xb7\x5a\xd6\xc2\x5c\x89\x17\x88\xf8\xd3\x12\x24\x6c\x37\xa2\x2d\x1c\x7c\x73\x05\xbd\x65\x7a\x36\xd2\x5b\x09\x2e\x21\x59\x3f\xa4\xdb\xb0\x72\x45\x6d\x6c\xe0\x1c\x36\x36\x77\xcc\xf0\xe3\xd9\xa5\xca\x3c\xd3\xda\xc4\xb9\x00\xda\x66\xb8\x2c\xc3\xc8\x0e\x8d\xac\x40\x8e\xee\xd2\xf5\x27\x28\xc0\xf0\xa8\x73\x3d\x98\x9b\x08\x72\xb2\xac\x5d\xef\x3f\xbe\x3e\x9b\xb3\x21\xa6\x48\x06\xed\x4f\x8d\x51\x5d\x63\x5e\x13\xdc\xa6\x82\xba\x85\xa0\x77\xfb\xd1\x15\x51\x2a\xa9\xd5\xb6\xc2\x5b\xf4\xf1\xf5\x7b\x7b\xfc\xc4\x0a\x2c\x9b\x63\x5c\x76\x0e\x69\x8d\x49\x1a\x63\x80\xde\x39\x3f\x1c\xc7\x6a\x9a\xb3\x61\x1b\x6d\x6a\x4e\xaf\x06\xd4\x6a\xed\x6a\xb4\x32\x6e\x72\x11\x2a\x46\xff\x25\x4f\x1b\x9e\x4d\x9a\xa3\xb1\x46\xa5\xf1\xcc\x2e\xd0\x3b\xc8\x25\x2a\x4a\x6c\xe6\x1b\xc7\x67\x7f\x08\xa3\x94\x6d\x5b\xbf\xda\x98\x6d\x8c\x3e\x73\xd8\xd8\xe1\x6d\x10\x00\x0a\x37\x0d\x69\xcc\x5a\x60\x3b\x88\x52\x04\xac\x1f\x8b\x55\xbf\xa8\xf3\xfc\x3b\x71\xb1\x74\x0b\x3a\xee\xff\xeb\xfe\x4d\x5f\xf1\x27\xf5\xfb\xe5\xf1\x7c\x2b\x60\x38\xb9\x06\x7c\x5b\x01\x95\x58\xa1\xaf\xd1\x0b\x74\xe5\xb9\xe7\x73\xc7\xb4\xf7\xfb\xa6\xbd\x2d\xe1\xd7\x6b\x80\x37\xa0\x73\x80\x73\xf4\xf7\xc1\x61\xeb\x52\x14\xa2\x7a\x69\xa1\x40\x08\xda\xfb\x2a\x5b\x8c\x03\x37\x88\xb5\xa8\x56\xc2\x54\x51\x59\x49\x1e\xaf\xc7\x8a\x18\x7c\xb2\x87\x82\xc3\xa3\xa4\x6d\x88\x2d\x00\x44\xa9\x88\xbe\x71\x9d\x41\xa4\x04\x5b\xf5\x98\xc1\x60\x15\x55\xe2\x91\xea\x92\xd2\x2c\x8e\x45\xb3\xc1\x34\x39\x03\x00\xee\x08\xdc\xc9\x55\x54\xee\x5e\xb4\xb7\xe0\xd7\x37\x21\xde\x49\x78\x13\xd9\x6e\xb7\x90\xdd\x48\xe6\xd0\xfc\x1d\xbb\x98\x28\xa9\x45\xed\x3d\x64\x6c\xcd\xb9\xc7\xf7\x2b\x02\xd3\x06\xdb\x19\x25\x42\x43\x84\x41\x6b\x76\x56\xf3\xe2\xc6\x21\x5f\xa2\x5b\x38\xd7\x0a\x1b\x1d\x54\x66\x0c\x71\x8a\xe0\x8e\x09\xa3\xbd\xcf\xd7\x26\xfb\x16\x98\x1a\x49\xed\x3d\x6c\x75\x1a\x45\xa8\x2b\x50\xa0\x50\x94\xde\xf2\x7e\xda\xf6\xc9\x9e\x33\xe1\x8b\xfe\x58\x26\xd5\xdf\x17\xdf\xff\xff\x6e\x6e\x4a\x7f\x2c\xe9\x78\x9f\xfa\x78\xa3\xd2\x98\xa5\xd0\xf1\x43\xf6\xce\x22\x82\x68\x40\x69\xe3\x9c\x5a\x2f\x17\x12\x9b\x5d\xb7\x12\x89\x29\x43\x58\x94\xa8\x7e\x2a\xc2\x58\x1d\x58\x61\x18\xab\xa5\xc9\x61\x3f\x2b\x62\xf1\xe4\x38\x16\x4e\xb7\x65\x65\x53\x1b\xdb\x72\x00\x24\x61\xd8\x2c\x31\x35\xb1\x45\x99\xc9\x27\xd6\xc7\xbc\x0c\x47\xba\xc6\x9a\xb5\xde\x50\xbf\xda\x81\xed\xee\x5e\x5d\xbc\x5c\x46\xfc\x87\x52\x7a\x99\xee\xf9\x67\x43\xb9\xf5\x34\x8c\xe3\xc3\x5b\x0b\x43\x3b\x34\x11\x70\xa4\xa0\x6e\x12\xff\xe1\xec\x00\xf7\x8b\x53\x75\x2f\x03\xb9\x8c\x00\xa2\x4d\xd8\x0f\xd3\xd1\xb1\x3d\x64\x27\x9f\x7e\xbe\x7e\x31\xd5\xc7\x76\x63\xad\x42\xcb\xc0\xbd\x78\x90\x31\x23\x8c\x05\x75\x17\x07\xdd\x40\xff\x53\x72\x03\xbb\x8d\x57\x7f\x15\x77\x40\xcd\xb4\x47\x3c\x26\x37\x1b\xc0\x37\x81\x82\x05\x0e\x00\x64\xcb\x5c\xa0\x89\xd4\xba\x88\xe8\x24\x16\x78\x30\xbe\xe5\x06\x85\x90\x76\xd4\x77\x54\xd5\x89\xb6\xdd\x57\x02\xe6\x32\x09\xfc\xde\x87\x28\xa8\x2c\x6b\x7d\xba\x9d\xff\x94\xab\xae\x7d\x0f\xfa\x4e\x4f\x66\x12\xa1\xd4\xba\x89\xd3\xac\xf7\x5b\x29\x4a\xe6\xb5\x11\xb0\x10\x78\x1b\x30\x6b\x95\x79\xaf\x30\x2f\x5e\xbd\x58\xea\x6e\x53\x7c\x8c\x55\x16\xf2\x81\x65\x60\xc9\xb3\xff\xf0\xa2\x32\x4f\x50\xba\xf3\x0e\x4f\x46\xe4\x8a\xdf\xac\x78\xe4\x84\x39\x80\x6f\x19\xc6\x16\x19\xe7\x59\x32\xef\xf2\xd4\x11\x66\xd2\xcb\x05\x94\xc2\x63\xbb\xe7\x7f\xdc\x14\x60\x37\x48\x76\x8b\xf0\x0d\xf5\xb6\xb7\x3f\x5b\xcc\x1b\xfa\x9f\x2b\xe8\x0d\x35\x52\x73\x84\x2f\x7c\xb8\xb9\x16\x55\x81\x80\xae\xc4\x73\x60\x6a\xc5\xe9\x98\x03\xf1\xe6\x3b\x06\x9f\x50\x14\xef\xa8\xa0\x0f\x0c\x89\x80\x24\xf1\x14\xe1\xdd\xc7\xa8\xa1\x28\xa1\x2e\xf5\xe8\x14\x29\xe0\xcb\xa7\x6f\xe5\x6a\x85\x23\x48\xe1\xa0\xed\x72\x77\x97\xe1\x53\x2c\x4a\x9d\xd2\x19\xe6\x1d\x36\x25\x0c\x32\x3c\xec\x52\x70\x1d\xbf\x78\x65\xbc\x43\x70\x90\xf8\x4d\x18\x5b\x7f\xfe\x68\x41\xf1\xfd\x07\x4e\x37\xe0\x1b\x6c\x66\x42\xd7\x05\x52\xa9\x38\x81\x5b\xd0\x48\x43\x7e\xcd\xb3\x33\xfd\xa7\x97\x84\x5d\x0c\xe8\xba\x7b\xe5\x6f\x57\x86\xee\x12\x8b\xc8\xfe\x74\x6d\xed\x92\x95\xb5\x3a\x1c\x01\x2f\x37\xcc\xd0\x20\x08\x06\x63\x98\x20\x35\xcf\xab\xae\xc3\xd5\x5c\xed\x16\x8c\xe3\xfc\xb5\x07\x5b\xed\xfd\x45\xa3\xe5\xfc\x37\x00\x00\xff\xff\x5c\x24\xc6\x2a\xe2\x10\x00\x00")

func data_publish_gh_pages_sh_bytes() ([]byte, error) {
	return bindata_read(
		_data_publish_gh_pages_sh,
		"data/publish-gh-pages.sh",
	)
}

func data_publish_gh_pages_sh() (*asset, error) {
	bytes, err := data_publish_gh_pages_sh_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "data/publish-gh-pages.sh", size: 4322, mode: os.FileMode(420), modTime: time.Unix(1424859334, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _data_srcco_css = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x9c\x56\x4f\x6f\xe3\xb6\x13\xbd\xe7\x53\x10\xf9\x61\x81\x64\x21\x29\xb4\x2d\xdb\x89\x7d\xf9\x6d\xd3\x2d\x7a\xe8\x16\xbd\xb4\x77\x4a\xa4\x2d\xd6\x14\x29\x90\x94\x6d\x6d\x90\xef\xde\x21\x45\xc9\xf2\x1f\x2d\x9a\x3a\x9b\x05\x4d\xce\x9b\xe1\xcc\x7b\x33\x4c\x61\x4b\x81\xde\xee\x10\x7c\x0a\xc6\xb7\x85\x5d\x21\x52\x5b\xb5\xf6\x3b\x25\x97\x71\xb7\x3b\xc1\xf8\xd3\xfa\xee\xfd\x2e\x53\xb4\x09\x80\xa7\xcf\x68\xa3\xa4\x8d\x37\xa4\xe4\xa2\x59\xa1\xfb\x5f\x99\xd8\x33\xcb\x73\x82\x7e\x67\x35\xbb\x8f\xfa\xef\xd1\x17\xcd\x89\x88\x0c\x91\x26\x36\x4c\xf3\xcd\x1a\x7d\x7e\xf2\x3e\xfe\xa5\x83\xf6\x3e\x19\xc9\x77\x5b\xad\x6a\x49\xe3\x5c\x09\xa5\x57\x48\x6f\xb3\x87\x17\x1c\xa1\xf6\xf7\xd1\x5d\x30\xd9\x6a\x4e\x2f\x52\x6a\x2f\xef\x53\x22\x7a\xcb\xe5\x0a\xe1\xea\x38\x48\xb4\x22\x94\x72\xb9\x6d\xb7\x27\xf3\xea\x38\x16\xef\x7f\xbf\x4c\xdd\x4f\x7b\x7c\xe0\xd4\x16\x5d\xbd\x20\xb0\x56\x87\x10\x57\xed\x99\xde\x08\x75\x58\xa1\x82\x53\xca\xe4\x45\x94\xc9\x79\x98\x4b\x3f\x54\xe5\xc1\x8f\xaf\x8e\xe1\xdf\x19\x60\x96\x9d\xf9\x87\x52\xe8\xee\x9d\xa6\x69\xbb\x01\xd7\x22\x50\x11\xc1\x36\xf6\x2c\xfa\x1c\x63\x87\x81\xf0\xb9\xa2\x6c\x18\xbf\x63\xe7\x1b\x93\x42\x45\xdf\x94\x24\xb9\x8a\x5e\x95\x34\x4a\x10\x13\xdd\xbf\xaa\x5a\x73\xa6\x81\xb1\xc3\x7d\x54\x2a\xa9\x4c\x45\x72\xb6\xbe\xba\xff\xb4\x4f\xb7\xe0\x96\xc5\xde\x6a\x85\x2a\xcd\xe2\x83\x26\xd5\xfa\x07\x75\xbb\x9d\x70\xa6\x8e\xb1\x29\x08\x75\xd6\xee\xc4\xe7\xed\xfe\x03\x4d\x90\x07\xd0\x43\xf8\x97\x4c\x1e\x11\x97\x86\xd9\x0e\xa6\x29\xd3\xb1\x26\x94\xd7\xa6\xa5\x62\x94\xec\xd9\x72\xf6\x32\x5b\x9e\x7a\x21\x94\x2a\xc5\x3d\xe8\x8c\xd2\xbe\x7a\xab\xc2\x25\x82\x48\xa8\x62\x7c\x60\xd9\x8e\xdb\x98\x48\x5e\x12\xcb\x15\xa4\xd2\x30\x01\x69\x42\x69\xa1\xd4\x38\x59\x18\xc4\x88\x61\x31\x84\x50\xb5\x45\x93\xd6\x77\x5c\xaa\xef\x1f\xc5\x7c\xc0\xbc\x63\x9a\x9c\xba\xf9\x94\x0d\x94\xd1\xf1\xd5\x75\x29\x9c\x75\x24\xc4\xee\x2c\xee\x0e\x87\x1a\xe3\xb2\x80\xd6\x0e\x65\xb6\xec\x68\x63\xca\x72\xa5\xc3\x75\xa4\x92\xec\x26\x03\xb3\x6b\x02\x40\x9e\x5c\x32\xa2\xe3\xad\x33\x62\xd2\x3e\x58\x05\x30\x6b\x55\x19\xb5\xf4\x4e\xe7\xf3\xa8\xfb\x4d\x66\xf3\x47\x84\x3f\x45\xd7\x07\x18\x0e\x5c\xeb\x3f\x0e\xb2\x0d\xdc\xbc\xfd\xe7\x90\x38\x4a\x96\xf8\x3a\x20\x8e\x70\x32\xc3\x83\x78\x77\x50\xb4\xb8\xff\x20\x49\xf6\x83\xaf\xb7\x3f\x50\xd0\xc4\xaa\xdc\x9c\x18\xa1\xdc\x90\x4c\x30\x68\x24\x0d\x15\x3c\x74\x7c\xc0\x76\x25\x48\x73\x5d\xd5\x15\x82\xae\x84\xf1\xe7\x48\xca\x04\x64\x77\x56\x72\xab\xaa\x21\xa4\x52\x86\xb7\xe4\x6c\xf8\x91\xd1\x6b\x16\x5a\x09\x5d\x37\xe1\x69\xbc\x78\x8f\xb8\x5d\xeb\x76\xd6\xe2\xb3\xb1\x32\x39\xf5\x4a\x37\x8c\xa7\x5d\xab\xb8\x54\x13\x92\x5b\xbe\xef\xe6\x4d\x00\xcd\xf0\xd0\x04\x9d\xdb\xf4\xb9\x67\x42\xb9\xfc\x5a\xab\x58\x92\x92\xdd\x60\x15\xe8\x6c\x46\x06\xdf\x30\x23\x7c\x31\x8b\xe7\xfd\x86\x17\x32\x11\x7c\x0b\x96\x39\xe8\x82\xe9\xf1\x6c\xfc\x25\x6e\x66\x34\x99\x0f\xcc\xc2\xd1\xa9\xfc\x24\x03\xd6\x6a\xcb\xce\x87\x60\x0c\x39\x9a\x5c\x2b\x21\xd6\x97\x8f\xd9\xc5\x6d\x67\xa7\x1d\x4f\xc8\xb4\xff\x7a\x21\x14\x88\x2f\xa1\x0b\xe2\xc1\x23\x7e\x6d\xe1\x6e\x28\x78\x38\x16\xdc\xc0\x04\xb7\x8d\x60\xb1\x6d\x2a\x36\x94\x4f\x5f\xbf\x0b\xb5\x23\xd3\x48\x4b\x8e\x30\xc2\xb7\x85\x70\x57\x86\x91\x72\x4b\xfa\xa0\x65\x00\xfd\xa1\x99\xb5\x0d\xbc\x03\x5c\x7a\x43\x1f\xcb\x24\xe8\x4f\xc3\x28\xe4\x67\x0b\xf7\x44\x58\xcb\x37\x4d\xf2\x37\x6c\xb7\xa0\xbf\x78\x89\x4c\x2d\xb3\x5a\x1b\x8b\x6c\xc1\x80\xfa\xac\x41\x3f\x93\x3d\x48\xff\x37\xc6\x33\xb5\xe7\xb9\x1f\x4f\x89\xb1\x3a\x42\xbe\xf3\xdd\x12\xbd\xf5\x0f\xe2\x62\xfe\x13\x4e\xa7\x6b\xf4\xee\xda\x0c\x8e\x5c\x6c\x14\x83\x5e\x18\x93\x0e\x9a\xec\x0e\xb4\x43\xc2\x72\x80\xfc\x3a\x7d\x7e\x59\xa4\x01\xb9\x63\xcd\x01\x9a\x0b\x90\x94\xe8\x1d\xaa\xb8\xdc\x79\x74\xee\xe6\x46\x8b\x86\xe5\x00\xfd\xe5\xab\xfb\x59\x87\xa7\xd1\x25\x0b\x93\xd3\x82\xc0\xf2\xe0\x11\xcc\x4b\xd0\x99\xbf\x0b\x69\xbc\x33\xa8\x7c\xe7\x0c\x96\x03\x67\xcf\x2f\x19\xdd\x6c\x02\xd0\xf1\x03\x28\x5f\x72\x68\x8e\x9a\x79\xac\xe0\xb6\xc3\xc2\x72\x80\x9d\xcd\x9e\x97\xaf\xaf\x01\x0b\x47\x4c\x13\x01\xf0\x1e\x58\xd5\xb2\x03\xc2\x72\x00\xdc\xf4\x11\x61\x3f\xb7\xb5\x1f\xf0\x80\xf4\xef\x7a\x0b\x15\x27\xa8\x18\x81\x0a\x02\x84\x43\x6f\x9d\x01\x2d\xd9\xf6\x89\x92\xed\x58\xa2\xee\x6f\xd6\xa7\x23\xfc\xdd\xea\x6c\xdc\x43\x79\x99\x33\xb1\x7d\x7c\x58\x0e\xdc\x64\x34\x5b\x2e\xb2\x4b\x37\xc4\x02\xfd\x19\xf4\x1f\xf2\x43\x04\xfc\xed\x0a\xb2\xe3\xc1\xd5\xfe\xe4\x6a\x3f\xa6\x9f\x1b\xae\xf6\xc4\xdd\x66\x28\x28\x78\x0d\x3b\x57\xb0\x1c\x63\x02\x8e\xe0\x09\x1f\x32\xf1\xff\x92\x51\x4e\xda\x0e\xf1\x5d\x39\x2e\x6a\xbc\xc0\xe0\xc6\x99\x8c\xaa\x17\xe3\x45\x10\xdf\x21\xcc\x93\x4c\x09\x1a\x50\xa3\xaa\x5d\x60\x3c\x22\x59\x07\x1b\xd5\x67\x8a\xd3\xf1\x60\xa3\xca\xc4\x69\x1a\x4c\x46\x35\x98\xa6\x5d\xa2\xa3\x5a\xc3\xb8\x33\x19\x55\xd5\x0f\x6b\x31\x2a\x22\x9f\x54\x30\x19\x11\x47\xe0\xe1\xfd\xee\x9f\x00\x00\x00\xff\xff\x37\x47\x9c\xa1\x5f\x0d\x00\x00")

func data_srcco_css_bytes() ([]byte, error) {
	return bindata_read(
		_data_srcco_css,
		"data/srcco.css",
	)
}

func data_srcco_css() (*asset, error) {
	bytes, err := data_srcco_css_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "data/srcco.css", size: 3423, mode: os.FileMode(420), modTime: time.Unix(1424859334, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _data_srcco_js = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xac\x54\x4d\x8f\xd3\x30\x10\xbd\xf7\x57\x18\x73\xa8\xab\x52\x8b\x7b\xe1\x80\x10\x37\x04\x07\xb8\x21\x0e\x5e\x7b\x9a\xb5\x70\x3d\xe0\xb8\x59\x55\xa8\xff\x9d\x99\xa4\x9b\x75\xb2\xcd\x36\x95\xf0\xa1\x1f\x9e\x99\xe7\xf7\xde\x8c\xfd\xe0\xa3\xc3\x07\x8d\x31\xa0\x71\xe2\xbd\xd8\x1d\xa2\xcd\x1e\xa3\x50\x2b\xf1\x77\x21\x68\x35\x26\x89\x68\xf6\x50\x53\xd4\xa1\x3d\xec\x21\x66\xfd\xe7\x00\xe9\xf8\x0d\x02\xd8\x8c\xe9\x43\x08\x4a\xea\x8c\x76\xc3\x79\x72\xb5\x6d\xeb\x76\x98\x84\xe2\x62\x4f\x85\x6f\xb7\xf4\xf5\xae\xc3\xd1\x01\x62\x95\xef\x69\x67\xbd\x7e\x3c\x84\x57\x17\xf4\x19\xf6\xca\xaf\xb4\x71\xee\x53\x43\x47\x7d\xf6\x75\x86\x08\x49\x49\x1b\xbc\xfd\x25\xdf\xf4\x14\x15\x34\x65\x39\xaf\x9c\x7c\x55\x41\xfa\xfe\xf5\x23\x05\x75\x36\xa9\x82\xac\xbd\x3b\x33\xe2\x75\x3a\xff\x3e\xf5\xda\xb2\xcf\xe1\xaa\xb8\x88\x0e\x36\x6d\xe6\x4b\xf2\x3a\xa8\x29\x7d\xe7\xe8\xff\x10\xf8\x85\xe8\x14\x0a\x7f\x9b\x44\x40\xbc\x79\x51\xe9\x69\xbb\x58\xf4\x7d\x2d\x2c\x22\x63\x8a\x1e\x27\xc4\x3c\xe9\x82\x92\xaf\xa5\x58\x0b\xef\xe8\x43\x6e\xa8\xd5\x8f\x36\xf8\x9d\x50\xaf\xb8\x54\xdb\x60\xea\x9a\xc5\x68\x8b\x31\x1b\x1f\x6b\x25\x0d\x9d\xd9\x90\x65\xa5\x8c\x51\x32\xd9\xf0\x94\xb7\x1d\x66\x3d\xe9\xba\xbd\x60\x44\xbf\x18\xcf\x99\x58\x63\x80\x76\x04\xee\xd0\x1d\xaf\x21\x9c\x04\x84\x1a\xa6\x15\x27\xd8\x63\x03\xb7\x89\x9e\x5f\x33\x4b\xf7\x34\x9c\x83\x76\xcf\x64\xe0\xc1\x9f\x34\x82\x5a\xda\x4f\x57\x31\x5c\xc3\x6a\x4e\x1c\xbc\x22\xf4\x9f\x2f\x5a\x7c\xc6\xb3\xb8\x64\x9d\xc3\x2f\x3c\x21\x0c\x32\xf9\x84\x70\xf0\x87\xff\x79\x55\xea\x90\x76\x79\xab\xc6\xa4\x99\xcf\x45\xce\x97\x08\xb7\xb7\x81\x37\x66\xdf\x06\x8a\xd6\x18\x40\x07\xac\xd4\xf2\xde\x2f\x8b\x56\x8c\x70\xe6\x4c\xda\x00\xed\xee\x08\xcb\xc9\xce\xb6\xe0\xcf\xde\x0d\xb6\xe5\x5f\x00\x00\x00\xff\xff\x18\xfb\x08\xc5\x12\x06\x00\x00")

func data_srcco_js_bytes() ([]byte, error) {
	return bindata_read(
		_data_srcco_js,
		"data/srcco.js",
	)
}

func data_srcco_js() (*asset, error) {
	bytes, err := data_srcco_js_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "data/srcco.js", size: 1554, mode: os.FileMode(420), modTime: time.Unix(1424859334, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _data_view_html = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x8c\x53\xbd\x4e\xc3\x30\x10\xde\x79\x8a\x23\x03\x5b\x93\x07\xc0\xe9\xd2\x82\x18\x40\xad\x68\x16\xc6\xd4\xbe\x34\x06\xd7\xae\x6c\x17\xa8\x22\xbf\x3b\xe7\x24\xad\x8c\x40\x81\xa9\xd7\xfb\x7e\xee\xf3\xe9\xc2\xae\x97\xab\x45\xf5\xb2\xbe\x83\xd6\xef\xd5\xfc\x8a\x0d\x3f\x00\xac\xc5\x5a\xc4\x82\x4a\x2f\xbd\xc2\x79\xd7\xe5\x55\x2c\x42\x60\xc5\xd0\x19\x50\x25\xf5\x1b\x58\x54\x65\xe6\xfc\x49\xa1\x6b\x11\x7d\x06\xad\xc5\xa6\xcc\x48\xf3\x8c\xce\x1c\x2d\xc7\x35\x35\xe4\x67\x08\xce\x72\x6e\x72\xee\x5c\x36\xea\x1d\xb7\xf2\xe0\x81\xfa\x13\xfc\x57\xa2\xb3\x62\xa0\xf6\xf1\x8a\x73\x3e\xb6\x35\xe2\x34\x5a\x09\xf9\x0e\x5c\xd5\xce\x95\x99\x37\xfc\x3c\x61\x04\xa4\x28\xb3\x46\x52\xc0\x2c\xe1\xcc\x74\xbd\xc7\x0b\x0f\xa0\x27\x9c\x55\x05\xc9\x7e\x58\x08\x6c\x26\x1d\x22\x3e\x69\xd0\x8f\x98\x91\x32\x75\x49\x0c\x68\x07\xf7\x44\xa9\xea\xad\xc2\x55\xb3\x30\xda\xa3\xf6\x2e\x84\x3f\x53\x4d\x7b\x6e\xbc\x3d\x72\x7f\xb4\x28\xfe\xe1\x9c\x96\xc9\x52\x77\x56\x8a\x8b\x6b\xd7\x81\xad\xf5\x0e\x21\xdf\xe0\x6e\xff\xdd\x29\xd1\x58\xf3\x91\x04\x49\x11\x11\x23\x76\x9d\x6c\x20\x5f\x1a\xfe\x50\x3d\x3d\x86\x40\x41\x93\x1a\x95\xa3\x73\xbb\xd1\x5b\x77\xb8\xa5\x7f\x5a\xc4\xdb\x4b\x9e\x1f\x43\x44\xf9\xc2\x08\x1c\x34\xa9\x3f\xa7\x6e\x1c\x90\xc2\xbd\x7a\xb4\xfa\x6d\xa1\xf4\x28\xc2\x60\x04\x2f\x10\x2b\x86\x2b\xa3\xb3\xeb\xbf\x8f\xaf\x00\x00\x00\xff\xff\xeb\x6c\x9b\x5d\x37\x03\x00\x00")

func data_view_html_bytes() ([]byte, error) {
	return bindata_read(
		_data_view_html,
		"data/view.html",
	)
}

func data_view_html() (*asset, error) {
	bytes, err := data_view_html_bytes()
	if err != nil {
		return nil, err
	}

	info := bindata_file_info{name: "data/view.html", size: 823, mode: os.FileMode(420), modTime: time.Unix(1424860011, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"data/publish-gh-pages.sh": data_publish_gh_pages_sh,
	"data/srcco.css": data_srcco_css,
	"data/srcco.js": data_srcco_js,
	"data/view.html": data_view_html,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() (*asset, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"data": &_bintree_t{nil, map[string]*_bintree_t{
		"publish-gh-pages.sh": &_bintree_t{data_publish_gh_pages_sh, map[string]*_bintree_t{
		}},
		"srcco.css": &_bintree_t{data_srcco_css, map[string]*_bintree_t{
		}},
		"srcco.js": &_bintree_t{data_srcco_js, map[string]*_bintree_t{
		}},
		"view.html": &_bintree_t{data_view_html, map[string]*_bintree_t{
		}},
	}},
}}

// Restore an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, path.Dir(name)), os.FileMode(0755))
        if err != nil {
                return err
        }
        err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
        if err != nil {
                return err
        }
        err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
        if err != nil {
                return err
        }
        return nil
}

// Restore assets under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        if err != nil { // File
                return RestoreAsset(dir, name)
        } else { // Dir
                for _, child := range children {
                        err = RestoreAssets(dir, path.Join(name, child))
                        if err != nil {
                                return err
                        }
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

