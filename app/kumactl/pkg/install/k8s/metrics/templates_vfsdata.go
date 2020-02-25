// Code generated by vfsgen; DO NOT EDIT.

// +build !dev

package metrics

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// Templates statically implements the virtual filesystem provided to vfsgen.
var Templates = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2020, 2, 25, 13, 21, 47, 966964578, time.UTC),
		},
		"/grafana": &vfsgen۰DirInfo{
			name:    "grafana",
			modTime: time.Date(2020, 2, 25, 9, 33, 42, 16258248, time.UTC),
		},
		"/namespace.yaml": &vfsgen۰FileInfo{
			name:    "namespace.yaml",
			modTime: time.Date(2019, 9, 9, 21, 11, 18, 649000000, time.UTC),
			content: []byte("\x0a\x2d\x2d\x2d\x0a\x61\x70\x69\x56\x65\x72\x73\x69\x6f\x6e\x3a\x20\x76\x31\x0a\x6b\x69\x6e\x64\x3a\x20\x4e\x61\x6d\x65\x73\x70\x61\x63\x65\x0a\x6d\x65\x74\x61\x64\x61\x74\x61\x3a\x0a\x20\x20\x6e\x61\x6d\x65\x3a\x20\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x73\x70\x61\x63\x65\x20\x7d\x7d\x0a"),
		},
		"/prometheus": &vfsgen۰DirInfo{
			name:    "prometheus",
			modTime: time.Date(2020, 2, 25, 14, 10, 9, 832318694, time.UTC),
		},
		"/prometheus/alertmanager.yaml": &vfsgen۰CompressedFileInfo{
			name:             "alertmanager.yaml",
			modTime:          time.Date(2020, 2, 25, 14, 10, 9, 817447599, time.UTC),
			uncompressedSize: 3970,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x57\x4d\x6f\x1b\x37\x13\xbe\xeb\x57\x10\x7a\xdf\x43\x7b\xd8\x5d\xc9\xaa\x13\x87\x80\x0f\xae\x8d\x1a\x06\x6a\x5b\xb0\xdb\xf4\x50\x14\xc1\x88\x3b\x92\x18\x73\x49\x96\x1c\x2a\xd9\xba\xfe\xef\x05\xf7\x43\xd9\x5d\xc9\x56\x50\x14\x81\x79\xd2\x72\x86\xf3\xf1\xcc\x33\x43\x2a\x49\x92\x11\x58\xf9\x1e\x9d\x97\x46\x73\xb6\x99\x8e\x1e\xa4\xce\x39\x3b\x37\x7a\x29\x57\xd7\x60\x47\x05\x12\xe4\x40\xc0\x47\x8c\x29\x58\xa0\xf2\xf1\x17\x63\xc2\x14\xd6\x68\xd4\xc4\xd9\x18\x14\x3a\x2a\x40\xc3\x0a\xdd\xb8\x92\x82\xb5\x9c\x59\x67\x0a\xa4\x35\x06\x3f\x62\x4c\x43\x81\xdd\xad\xa4\x7b\xa8\x91\x7b\x0b\x02\x39\x7b\x7c\x64\xe9\x4d\xfb\xc9\x9e\x9e\x46\xad\xff\xee\x91\xb4\x2c\x14\x67\x7f\x57\xde\x56\xca\x2c\x40\x71\xf6\xf8\x54\x7d\x3a\x14\x28\x37\xe8\x9a\x48\x93\xc6\x77\x8e\x4b\x08\x8a\x92\x56\x5c\xeb\x9a\x40\x58\xeb\x31\xb6\x72\x26\xd8\x0f\x52\x13\xba\x4d\xb4\x77\x5c\xf4\x04\x9f\x40\x12\x67\xd3\x89\x6f\x76\x5b\x43\xcf\x98\x8e\x0a\x16\x81\x3a\x06\x67\xeb\xd1\xb3\x90\xcf\xe3\x8e\x27\xd4\xf4\xde\xa8\x50\xe0\xb9\x02\x59\xbc\x0a\xf8\xbd\x45\x51\xc1\x2f\x04\x7a\x7f\x6d\x72\xdc\x22\x7b\x87\x90\xff\xe6\x24\xe1\xad\x16\x38\x8a\x19\x7b\x13\x9c\x68\x15\x1c\xfe\x19\xd0\x93\x6f\x01\xf6\x64\x1c\xac\x90\xb3\xf1\xd1\xa5\x1c\x3f\x8f\xc5\x3d\xba\x8d\x14\x78\x26\x84\x09\x9a\x5e\x05\x08\xc3\x58\xdd\x02\x44\x0a\x81\xd6\xc6\xc9\xbf\x80\xa4\xd1\xe9\xc3\x89\x4f\xa5\xc9\x36\xd3\x05\x12\x6c\x3b\x49\x05\x4f\xe8\xee\x8c\xc2\x6f\x94\x87\x0b\xaa\xc6\xff\xf7\x3f\xfe\x8b\xa8\x7f\x94\x3a\x97\x7a\xf5\x8d\x82\xf7\x61\xf1\x11\x45\xcd\x98\x84\xed\x65\x43\x34\x7b\xa8\x92\x2f\xd7\xd2\x19\x85\x77\xb8\xac\x38\x6d\xe5\x65\x6c\xee\x17\xa0\x19\x31\xb6\x5b\xca\x43\x21\x1c\xa2\xf6\xab\xe0\x74\xdb\xd8\xd6\x38\x1a\x0c\xcb\x35\x91\x6d\x9a\x36\x4a\x39\x3b\x99\xb4\x9f\xce\x90\x11\x46\x71\xf6\xcb\xf9\xbc\xd9\x23\x70\x2b\xa4\x79\xa5\xf8\x6e\xf2\x6e\x36\x62\xcc\xa3\x42\x41\xc6\xed\x64\xb5\x53\xa9\xdd\x9c\x3c\xfa\x88\xda\xd9\x72\x29\xb5\xa4\x92\xb3\x1b\xa3\x23\xe4\x54\xda\x38\x3b\x9a\x3a\x5c\xcd\xeb\x09\xf2\x3f\x76\x5f\xcd\x9c\xae\x91\x8c\xb0\xb0\x0a\x08\x7d\xd6\xf5\x97\xe4\x68\x95\x29\x0b\xd4\x94\x96\x50\xa8\x5e\x89\xc0\x5a\x9f\x6d\xeb\x74\xb1\xd5\x7c\x55\xa5\xea\xc3\x5a\x00\x89\xf5\xcf\x9d\x90\x0e\x05\xb5\x2f\x2c\x87\x56\x49\x01\x9e\xb3\x69\xc4\xb8\x01\xae\x71\xd0\x49\x3d\x2e\xd5\xf3\x75\xd8\xdb\x3e\x7f\x8c\xb5\xc9\x54\xbf\x7b\xcd\x7d\x73\xb0\xaf\xa3\x4f\x4d\x20\xf5\xf6\x7e\xef\xd2\xf6\xe5\x93\x71\xc9\xa2\xbe\x80\xa2\x66\x8f\x1b\x7c\x33\x49\x8f\x26\xe9\x64\x3c\x54\x9e\x07\xa5\xe6\x46\x49\x51\x72\x36\xbe\x5a\xde\x18\x9a\x3b\xf4\xa8\xa9\xab\x89\x7a\xc3\x3b\x9f\x5f\x22\x9a\xdf\x5e\x7c\xb8\x9a\xf7\x44\x8c\x6d\x40\x05\xfc\xc9\x99\x82\x0f\x04\x8c\x2d\x25\xaa\xbc\x19\x4e\xc3\x35\x18\x28\xbb\x0a\xd5\xe1\x39\xd0\x9a\x33\x4f\x40\xc1\xa7\xd6\xe4\x3d\xef\xe0\x56\x7e\x18\x68\x92\x88\xea\xb9\x97\x2e\xa5\xc2\xd3\x0c\x49\x64\xf5\x46\x36\x7c\x70\xed\x1c\x6c\x2e\xf4\xd4\x02\xad\x4f\xb3\xc8\x94\x5d\xdb\x75\xb7\xa6\x90\x6f\xd0\x91\xf4\x98\x40\x9e\x3b\xf4\xfe\xf4\xff\xdf\xd5\xe0\x7c\xcf\xdf\xbc\x3d\x99\xed\x1c\xfc\x84\x8b\x14\x3f\x13\x3a\x0d\x2a\x09\x4e\x9d\xc6\x99\xc4\xb3\x4c\x19\x01\x6a\x6d\x3c\xf1\x66\xd4\xb4\xab\x33\xc6\xbe\xd8\xd9\xb2\xa5\x37\x9d\xda\xe5\x10\x72\xa9\xd1\xfb\xb9\x33\x0b\xec\x9f\x8d\xee\x2e\x91\x86\x85\xb0\x15\xbc\x59\x92\xc5\xb3\xe5\x50\xb8\xcf\x09\x63\x71\x8c\x49\x50\x17\xa8\xa0\xbc\x47\x61\x74\xee\x39\x9b\x4d\x7a\x3a\x24\x0b\x34\x81\xf6\x8b\x07\xef\xaa\x76\x35\x0f\xde\x7a\x6d\xaa\x67\xe3\x75\x6c\xa3\x1d\x14\x6a\x2e\xd6\x55\x4d\x6a\xc5\x41\xe4\x45\x3c\x57\x53\xa7\xc3\x80\xbd\x66\x9a\xa2\x1f\xb4\x33\xae\x08\x31\x1e\x28\xf8\xb0\x68\xc4\xe3\xaf\x6d\xdf\x86\x9f\x05\xd8\xc4\xa1\x32\x90\xef\xe9\xe7\x8f\xb2\x28\x64\x5e\x7a\xa3\xb3\xa1\x76\xec\xec\xd9\xbf\x6d\xec\xfd\x0d\x53\xa7\x9e\xe4\xd2\x9d\x3e\x8f\x56\x45\xe1\xb5\x31\x0f\x5d\xf6\x4e\x8f\xde\xa6\x93\x74\x92\x4e\x2b\xf6\x56\x34\x1a\x64\xf4\x2a\x6a\x5d\xb7\xc6\xad\x56\x25\x67\xe4\x02\x6e\xa7\xb5\x08\x4e\x52\x79\x6e\x34\xe1\xe7\x4e\x6f\x2c\x7d\xf3\x8c\x7a\x73\x7c\x3c\xfb\x61\xbb\xed\x82\x3e\x7b\x41\x72\x63\xf4\x9d\x31\xd4\x73\xd1\x88\x7e\xf5\xf1\x8f\x55\xf7\x4c\x9d\xca\x9e\xa1\xff\x5c\xaa\xa2\xfd\x13\xdb\x47\xe8\xeb\x6e\x8a\x83\x64\xb7\xfb\xfe\xaf\xf5\x3d\x89\xb8\xf5\xf2\x95\xf6\x4f\x00\x00\x00\xff\xff\xc1\x9b\xed\x28\x82\x0f\x00\x00"),
		},
		"/prometheus/kube-stats-metrics.yaml": &vfsgen۰CompressedFileInfo{
			name:             "kube-stats-metrics.yaml",
			modTime:          time.Date(2020, 2, 25, 14, 10, 9, 832244573, time.UTC),
			uncompressedSize: 3770,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x56\x4b\x8f\xdb\x36\x10\xbe\xfb\x57\x10\xbe\xdb\x5b\xa3\x4d\x91\xea\x96\x6e\x81\x22\x40\xb1\x30\xb6\x8f\xfb\x88\x1a\xcb\x8c\x29\x0e\x77\x38\xf4\xc6\x5d\xec\x7f\x2f\xa8\x87\x25\xd9\xeb\xd8\xb5\x8b\x26\x3a\x89\xe4\x90\xdf\xc7\x6f\x1e\x9c\xd9\x6c\x36\x01\x6f\xfe\x42\x0e\x86\x5c\xa6\xb6\x8b\xc9\xc6\xb8\x22\x53\xbf\x23\x6f\x8d\xc6\x0f\x5a\x53\x74\x32\xa9\x50\xa0\x00\x81\x6c\xa2\x94\x85\x1c\x6d\x48\x7f\x4a\x69\xaa\x3c\x39\x74\x92\xa9\xe9\x26\xe6\x38\x0b\x02\x82\xb3\x0a\x85\x8d\x0e\xd3\xda\x06\xbc\xcf\x94\x67\xaa\x50\xd6\x18\xc3\x44\x29\x07\x15\x0e\xa7\x66\xc7\x5b\x5b\xab\xe0\x41\x63\xa6\x5e\x5e\xd4\xfc\xa1\x1b\xaa\xd7\xd7\xc9\x21\x6f\xce\x41\xcf\x21\xca\x9a\xd8\xfc\x0d\x62\xc8\xcd\x37\xef\xc3\xdc\xd0\xdd\x76\x91\xa3\x40\x77\xad\x7b\x1b\x83\x20\x3f\x92\xc5\xaf\x70\x27\x8e\x16\x6b\x90\x99\x02\x6f\x7e\x65\x8a\xbe\xc5\x4c\x53\xd3\xe6\x68\xc6\x40\x91\x35\x0e\x56\xf6\x4a\x84\x7e\x8a\x8a\xc1\xc8\x27\x1d\x82\xa0\x93\x2d\xd9\x58\xa1\xb6\x60\xaa\xc1\x32\x15\xfd\x20\x34\x9e\xed\x27\x3a\xc0\xa7\x48\x02\xc3\x69\x6f\x8d\xae\xb5\xd4\xe4\x84\xc9\x5a\xe4\x7e\xd9\x9a\xca\x08\x83\x2b\xff\x05\x8d\x83\xe5\x7e\x05\x5d\xe1\xc9\x38\x19\xb2\xd4\x8c\x83\xb1\x26\xb7\x32\x65\x05\xbe\x99\xda\x22\xe7\x03\x81\xac\x09\xb2\x1f\x3c\x83\xe8\xf5\x29\x91\xf1\xb3\xa0\x4b\x31\x13\x4e\x89\x5d\x00\x56\xe4\xc2\x10\xbc\x40\x6f\x69\x57\xe1\x90\xa0\x71\x25\x63\x08\x78\x24\xd8\x7e\xe7\xd5\x1c\xc1\xfb\x5b\xd9\xd5\x71\xb7\x8a\x76\x64\x78\x86\x60\x89\x72\x05\xd9\xbc\x5d\x7c\x93\xad\x66\x72\x9f\x28\xef\x29\xec\x07\xd7\x8b\x13\x85\x82\x06\x6b\x5c\x79\x0a\xb5\x2e\x02\xe4\x04\xac\xa7\xa2\xb3\xef\x62\xf7\x6a\x60\x4f\xd6\xe8\xdd\x29\x4c\x4f\x45\x61\x02\x47\x9f\xf2\x25\x8f\x45\x79\x73\x14\x04\x21\x86\x12\xdb\x3a\x76\x0a\xb7\xb5\xd2\x16\x46\xb1\xd8\x24\x18\x88\x80\x5e\xf7\x91\x71\x35\x17\x8d\x2c\x66\x95\xaa\x01\x86\x33\x84\x86\xa6\xa6\x74\x75\x9a\x3c\x45\x0c\xb7\x52\x70\x28\xcf\xc4\x1b\xe3\xca\x33\x04\x5a\xc3\xda\x5d\x06\x6f\x8d\xb6\xa2\x32\x21\x95\x0b\xc6\xd2\x04\xe1\xe1\xd3\x72\x8a\xc0\x16\xac\x29\x40\x8c\x2b\x9f\x31\x5f\x13\x6d\x9a\xda\x15\x9b\xcd\xbd\x8f\xaa\x28\x67\xac\xce\xd3\xfe\x0f\xde\xc1\x9f\x8d\x2b\x52\x32\xfd\xff\xcf\x61\x88\xf9\x27\xd4\xd2\xbe\x88\x6f\xf6\x1d\xe9\xf0\xcb\xba\x85\x2f\xf7\x0b\x4c\x16\x1f\x71\x95\x90\x3a\x2f\x7f\x41\xac\x89\x52\xc7\xed\xc2\x65\x44\xce\x35\x54\x23\x99\xc1\x39\x92\xc6\xdf\x8d\xd6\xfd\xd9\xc9\x65\x41\x33\x78\xcc\xd4\x54\x38\xe2\xf4\x1b\xe9\xbc\x82\x47\x9d\x18\xe8\x46\x9a\x8f\xcb\x4c\x3d\x90\x4b\xf2\x78\x62\x69\xc9\xcd\x5a\xa0\xb5\x88\x6f\x23\x36\xad\x66\xea\xfd\x77\xdd\x90\x49\x48\x93\xcd\xd4\x1f\xf7\xcb\x76\x4e\x80\x4b\x94\x65\x6b\xd8\x9a\x76\x47\x09\x5a\x4c\x0c\x77\xe3\xf3\x16\x97\x9f\x97\x4c\x03\x5a\xd4\x42\x7c\x8b\x86\xb2\xab\x9d\x72\xdf\x09\x30\x3d\x72\x7a\x7a\xc0\xef\xf6\x9e\xff\x65\xff\x3e\x7f\x73\x6d\x74\xe7\xcc\xb1\x2c\x55\x2a\x2d\xbf\x0d\xe8\x5d\x46\xf0\x2d\x8a\x5d\xcb\x91\xa9\xa4\xbe\x60\xe5\x2d\x08\xb6\x30\x03\x31\xd2\x67\x47\x88\x97\x62\xbe\x85\xaa\x54\x77\xb1\xfa\x7f\x54\x52\x1e\x2e\xac\x26\x09\xdf\x09\x18\x87\x3c\xe0\x34\xbb\xb8\x1a\x35\x9f\xa9\xa0\x4c\xc1\xf2\x14\x61\x97\x52\x5a\x13\x23\x85\xbb\xe3\x3d\xd9\x76\x31\xff\x69\xbe\x98\x1e\xee\x5d\x46\x6b\x97\x75\xe7\x91\xa9\xe9\xc7\xd5\x03\xc9\x92\x31\xa0\x93\xa1\xe5\x20\xf3\x0e\x89\x1e\x53\x1a\x5d\xed\x20\xd9\x0e\x77\x1f\x26\xdd\xe9\xfd\x8b\x81\xc5\xd1\x9b\xd8\x7c\x2f\xaf\x7b\x87\xe8\xc8\x46\x76\xf7\xe4\x04\x3f\x4b\x6f\xc5\xd1\x7d\x08\x0f\xe4\x1e\x89\x24\x53\xa9\xea\x8d\x97\xfe\x0c\xc8\x99\xfa\xf1\xdd\xbb\xef\x7f\x98\xfc\x13\x00\x00\xff\xff\x5b\x5b\x7f\x72\xba\x0e\x00\x00"),
		},
		"/prometheus/node-exporter.yaml": &vfsgen۰CompressedFileInfo{
			name:             "node-exporter.yaml",
			modTime:          time.Date(2020, 2, 25, 14, 10, 9, 823821441, time.UTC),
			uncompressedSize: 1824,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x94\xcf\x8e\xd3\x30\x10\xc6\xef\x79\x8a\x51\xee\x69\xb7\x37\xb0\xc4\x01\xed\x5e\x56\x82\x12\x51\xe0\x3e\xb8\xb3\x6d\xb4\x8e\xc7\xb2\x27\x85\xa8\xea\xbb\x23\xe7\xcf\xd6\x69\xb3\x14\x71\x01\x9f\xe2\xf1\xe7\x99\xdf\x7c\xb6\x53\x14\x45\x86\xae\xfa\x46\x3e\x54\x6c\x15\x1c\x56\xd9\x73\x65\xb7\x0a\x36\xe4\x0f\x95\xa6\xf7\x5a\x73\x63\x25\xab\x49\x70\x8b\x82\x2a\x03\x30\xf8\x9d\x4c\x88\x5f\x00\x9a\x6b\xc7\x96\xac\x28\xc8\x2d\x6f\xa9\xa0\x9f\x8e\xbd\x90\xcf\xbb\x65\x74\x4e\x81\xf3\x5c\x93\xec\xa9\x09\x19\x80\xc5\x9a\xd2\x50\x31\xd9\x35\x08\x82\x43\x4d\x0a\x8e\x47\x58\xac\xc7\x29\x9c\x4e\xd9\x2d\xda\x09\x26\x5a\xcb\x82\x52\xb1\x1d\x58\xcf\x45\x17\x15\x2f\x83\xf6\xe8\x48\x41\x2e\xbe\xa1\xfc\x9f\xb6\x15\x1c\xe9\x58\x57\x9b\x26\x08\xf9\xc7\x52\xc1\x9a\x2d\x65\x00\x71\xfb\x80\x54\x0c\x35\x6a\x12\x5f\xe9\xd0\xc5\x7a\x81\x82\xb7\xab\xbb\xbb\x31\xe0\x59\x58\xb3\x51\xf0\xe5\xbe\x1c\x62\x82\x7e\x47\x52\xa6\xd2\x40\x86\xb4\xb0\xff\xcb\x76\xa5\xed\xac\xbb\x1f\x81\xf3\xab\xa3\x41\xe7\xc2\xf2\xe5\x7c\x1e\x90\x6a\xb6\x1b\xfa\x5f\x2e\xd2\xe8\xf8\xd4\x86\x1a\x45\xef\x3f\x24\x50\x37\xb1\xe6\xc0\x1a\xb7\x45\xa1\x8d\x78\x14\xda\xb5\x7d\xa2\xde\xaf\xcf\x6c\x4c\x65\x77\x5f\x3b\x41\x74\x91\x6a\x67\x50\x68\xa8\x9e\x38\x13\x87\x99\x80\xfc\x01\xca\x1c\x0c\xc0\xd8\x6a\xf7\x3d\x79\xd4\xeb\x5b\x0e\xf6\x55\xad\x60\x65\xc9\x27\x24\xc5\x6d\xf3\xc7\x51\xd5\xb8\x8b\x37\x25\x4a\x97\x13\x91\x3a\xdc\x2d\x56\x6f\x16\xab\xfc\x52\x5d\x36\xc6\x94\x6c\x2a\xdd\x2a\xc8\x1f\x9f\xd6\x2c\xa5\xa7\x40\x56\x52\x25\xfa\x5d\x02\xd4\x43\x15\x85\x43\xd9\x2f\x9c\x67\xfd\x14\xde\x2d\xf7\x1c\x64\x19\x27\xf3\xb2\xd0\x86\x17\x55\x68\x43\x22\x4a\x1e\xdd\x65\xc7\xd3\xc7\x77\x3e\x99\xc1\xa3\xf2\xf2\x35\x8e\x23\x56\x99\x5d\xf4\x14\xb8\xf1\x9a\x2e\xea\x1d\x4f\xc9\xf4\xc0\xa6\xa9\xe9\x63\x3c\xb2\x57\xb0\xae\xba\x04\xa8\xa3\xbc\x44\xd9\x2b\x78\xc5\x89\x58\x1c\xb7\x9f\xac\x69\x15\x40\xfc\x09\xce\xa6\x9e\x5a\x33\x9b\xf9\x5a\x72\x4e\x9c\xe4\x8d\xda\x35\xc9\x0f\xf6\xcf\x57\xf1\xf2\xf1\x61\x12\xeb\x5b\x9e\xbf\x72\x69\x17\xdd\xde\x88\x32\x01\x70\x3d\xdc\x44\x3b\xdf\xcf\x6f\xf7\x47\xe9\xaf\x00\x00\x00\xff\xff\x13\x42\xe4\x8c\x20\x07\x00\x00"),
		},
		"/prometheus/pushgateway.yaml": &vfsgen۰CompressedFileInfo{
			name:             "pushgateway.yaml",
			modTime:          time.Date(2020, 2, 25, 14, 10, 9, 820588854, time.UTC),
			uncompressedSize: 2197,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x54\x41\x6f\xdb\x3c\x0c\xbd\xfb\x57\x10\xb9\x3b\x69\xf0\x7d\x1d\x56\xdd\xba\x16\x28\x0a\x0c\x81\xd1\x6e\xbb\x0c\x3b\x30\x0a\x1b\x6b\x95\x45\x41\xa2\xb2\x65\x45\xff\xfb\x20\xd7\x69\xed\x34\xa8\x81\x62\xe8\xa6\x93\x4c\x52\x7c\x8f\x7c\xa4\xcb\xb2\x2c\xd0\x9b\x2f\x14\xa2\x61\xa7\x60\x33\x2f\x6e\x8d\x5b\x29\xb8\xa6\xb0\x31\x9a\x4e\xb5\xe6\xe4\xa4\x68\x48\x70\x85\x82\xaa\x00\xb0\xb8\x24\x1b\xf3\x0d\x40\x73\xe3\xd9\x91\x13\x05\x13\x9f\x62\xbd\x46\xa1\x1f\xb8\x9d\xb4\x4e\xf4\x5e\x81\x0f\xdc\x90\xd4\x94\x62\x01\xe0\xb0\xa1\xbe\xa9\xec\xbd\xe9\xdc\xd1\xa3\x26\x05\x77\x77\x30\x5d\xec\x3e\xe1\xfe\xbe\xd8\x67\x1a\x96\xa8\xa7\x98\xa4\xe6\x60\x7e\xa1\x18\x76\xd3\xdb\xf7\x71\x6a\x78\xb6\x99\x2f\x49\x70\x57\xc8\x99\x4d\x51\x28\x5c\xb1\xa5\x37\xa9\x22\x24\x4b\x6d\xda\xaf\xdf\xfe\x04\xe7\x0f\xc6\xad\x8c\x5b\xbf\x09\xf5\x98\x96\xdf\x49\x4b\x9b\xb9\x84\x83\x73\x90\xb3\x8e\x88\xf8\xb2\x8c\x81\x2d\x5d\xd1\x4d\x86\x40\x6f\x2e\x02\x27\xff\x42\x5f\x0a\x80\xe7\x2a\x8e\x30\x18\x1b\xe9\x41\x2b\xd1\x39\x96\x16\xb1\xeb\xe7\x53\xd2\x2c\x8b\x0f\xbc\xcc\x48\x83\x02\xff\xce\xfc\x47\x4f\x3a\x63\x7a\x0e\xd2\x81\x97\x5d\xbe\x5a\xc4\xb7\x86\x07\xaf\x82\x93\xa3\x93\xf9\xce\x10\x58\x58\xb3\x55\xf0\xe9\xac\xea\x6c\x82\x61\x4d\x52\xf5\x43\x23\x59\xd2\xc2\xe1\x55\x55\xc9\xd6\x93\x82\x49\xa7\xd1\x65\x35\x79\xa6\x01\x7a\x1f\x67\x8f\x42\x9c\x93\xb7\xbc\x6d\xe8\x9f\xf8\xaf\xec\xfa\x3a\xec\x40\x83\xa2\xeb\x8f\x3d\x42\x23\x94\x0e\x91\x0a\xe4\xad\xd1\x18\x15\xe4\x0e\x0b\x35\xde\xa2\x50\x97\xbf\x57\x77\x3e\x76\x00\x35\x0a\x76\x08\x0e\x60\x57\x4a\x7b\x1f\xec\xed\x62\x6c\x65\x33\xa2\x13\x34\x8e\x42\x8f\x45\x39\xbe\xeb\x0f\xc7\x34\xb8\xce\x23\x90\x03\x67\xbd\x10\xb5\x99\x4f\x8f\xa6\xf3\xc9\x7e\x68\x95\xac\xad\xd8\x1a\xbd\x55\x30\xb9\xbc\x59\xb0\x54\x81\x22\x39\xe9\x47\x62\x58\xf7\xb8\x0c\x26\xff\x89\xe0\x23\xed\x6a\x7f\xf2\xdb\xb6\x9a\x0d\x39\x8a\xb1\x6a\xf7\x78\xf0\x34\x2f\xcd\x05\xc9\xd0\x08\xe0\x51\x6a\x05\xb3\x72\x56\x13\x5a\xa9\xb7\xfb\xee\x43\x28\x00\xc6\x19\x31\x68\xcf\xc9\xe2\xf6\x9a\x34\xbb\x55\x56\xfd\x68\x10\x23\xa6\x21\x4e\x72\xd8\x1d\x08\x57\xe6\x95\x4c\xf3\xdb\xb7\xe3\x19\x39\x05\x4d\x7b\x3a\xdc\xdd\x3f\x8e\x9d\x4e\xc1\xc8\xf6\x8c\x9d\xd0\xcf\x1e\xe7\x90\xdc\x69\x5c\xb0\xbb\x62\x16\x05\x12\x12\x0d\x5d\x9f\x23\x05\x05\xef\x8e\x8f\xff\xfb\xbf\xf8\x1d\x00\x00\xff\xff\x48\xf5\x8d\x19\x95\x08\x00\x00"),
		},
		"/prometheus/server.yaml": &vfsgen۰CompressedFileInfo{
			name:             "server.yaml",
			modTime:          time.Date(2020, 2, 25, 14, 10, 9, 828501864, time.UTC),
			uncompressedSize: 13174,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x5b\x5f\x6f\xdc\x36\x12\x7f\xf7\xa7\x20\x16\x79\x48\x2e\x90\xe4\xb5\xe3\xbb\x54\x41\x10\xe4\x52\x20\x28\xd0\xb8\x0b\xf7\xcf\x3d\xf4\x7a\x02\x45\xcd\xee\x32\xa6\x48\x1e\x49\x6d\xbc\x97\xcb\x77\x2f\x28\x51\x5a\xea\xaf\xbd\xee\x3a\x4e\x53\xfb\xa5\x59\x72\xc8\xf9\xcd\x70\xe6\xc7\x21\xc5\x06\x41\x70\x84\x25\xfd\x05\x94\xa6\x82\xc7\x68\x33\x3f\xba\xa4\x3c\x8b\xd1\x1b\xc1\x97\x74\xf5\x0e\xcb\xa3\x1c\x0c\xce\xb0\xc1\xf1\x11\x42\x0c\xa7\xc0\xb4\xfd\x17\x42\x44\xe4\x52\x70\xe0\x26\x46\x33\x0d\x6a\x03\x6a\x56\xb6\x63\x29\x63\x24\x95\xc8\xc1\xac\xa1\xd0\x47\x08\x71\x9c\x83\xdf\x14\x54\xe2\xae\x47\x4b\x4c\x20\x46\x1f\x3f\xa2\xf0\xbc\xfe\x89\x3e\x7d\x3a\xaa\x75\x62\x06\xca\x50\xbe\x4a\x54\xc1\x40\x87\xdb\x9c\xc5\xe8\xff\xa5\xa6\x8f\x9f\xea\x6e\xdd\x6a\xda\x69\xf2\xa5\x57\x4c\xa4\x98\x55\xd8\x11\x82\x0d\x66\x05\x36\x54\xf0\x84\x72\x03\x6a\x83\x59\x8c\xe6\xb9\xeb\xd5\x44\x61\x09\x13\x3d\x86\xe6\x20\x0a\x13\xa3\xf9\xb1\x2e\x7b\x2c\xb8\x64\x49\x19\x38\xef\x04\x28\x02\x43\x22\x52\xba\x31\x52\x40\x84\xca\x5a\x46\x0c\x48\xf5\x2d\x1d\x9a\xca\xf6\x8d\x0d\xae\x3a\x1c\xc6\xaa\xa3\xc1\xf3\x5e\xa4\x49\x77\x25\x6a\xa3\x0c\x36\x94\xb4\x07\xd8\x21\x06\xab\x15\x98\xa6\xc1\x36\x31\x41\x30\x5b\x0b\x6d\xe2\x6f\x8e\xbf\x39\x76\x53\xa7\x80\x15\xa8\xc4\x88\x4b\xe0\xa5\x13\x62\x14\x6d\xb0\x8a\x54\xc1\x23\x0d\x44\x81\xd1\xd1\x65\x91\x82\xe2\x60\x40\x87\x54\x44\x36\x02\x28\x01\x4c\x88\x28\xb8\x89\xca\x91\x4e\xcb\x0e\xe7\x6e\x48\x80\x25\xad\x82\xa6\x86\xbc\xeb\x4b\x74\xd6\x47\xae\x84\x05\x01\x3c\x93\x82\x72\x53\x0f\x52\x50\x06\x70\x5f\x1c\x13\x53\x86\xff\x25\x80\x6c\x8c\x55\xb0\x82\xab\x18\x65\xb0\xc4\x05\x33\x2f\x76\x1a\x5f\xac\x8d\x91\xba\x91\xd3\xa2\x50\x04\x12\x3f\x37\xaa\x59\x93\xc4\xe6\x4e\xe2\x41\x6d\xe2\x7d\x52\xca\x39\xa7\x94\x9e\x14\xac\xed\x4b\xa4\x50\xc6\x17\xd7\x64\x0d\xd6\x83\x3e\x50\xc3\xb4\x33\x7c\x87\x91\xe0\x5b\x2d\x17\xc1\x21\x51\xa6\x99\x85\x72\x0d\xa4\x50\x90\xe8\x4b\x2a\x93\x0d\x28\xba\xdc\xc6\xc8\xa8\x02\x3e\x5b\x7c\x70\x91\xc1\x5e\xa1\x61\x07\xdc\x34\x2a\xca\xee\x1c\xf7\x22\x63\x60\x7d\x45\xe6\x02\x21\x79\x1c\x3e\x7d\xb2\xd3\x09\x92\x61\x02\x79\xc9\x96\x9e\xa5\x2e\xb6\x42\xbd\x21\xf1\xb3\x67\xa7\x8d\x82\x2a\xf1\xaa\x99\xac\x1e\x9c\x65\x0a\xb4\x4e\x12\x6f\xc6\x12\x82\xa7\x05\xb5\xb5\x44\x58\xd2\x68\x33\x8f\x4a\xcf\x44\x8f\xe6\x91\x54\xe2\x6a\x1b\xe5\x60\x14\x25\xb7\x0b\x5e\x6b\x5c\x2b\x26\xbb\x30\xdd\xe4\x89\xc4\x66\xdd\x60\xfd\x6b\xc6\x62\x40\x70\xb6\xa1\x5a\xa8\x87\xa0\xbc\x71\x50\x46\x1d\x9f\xfd\x05\xa3\x73\x30\xa0\xdc\x9c\x41\x77\x3b\xfb\x7c\x7b\x60\x83\x70\xcf\x45\xa9\x77\x32\xcc\xb9\x30\x55\xb5\xb5\x2b\x3e\x12\x2a\x92\xaa\x54\xe9\x61\x70\x41\xd3\x85\xf1\xb8\x5c\xa5\x57\x4f\xee\x08\x8a\x8d\x84\xd1\xe8\xa9\xba\xbd\x60\xbf\x0e\xab\x9f\x04\x87\xc4\x69\xc3\x77\xbf\x18\xbf\x1e\xeb\xaf\xff\x89\x7f\x7b\xfa\xe4\xf1\xab\x38\xfe\x77\xf6\xf4\xc9\xab\x17\x8f\xed\x7f\x86\x53\xf8\xd1\x3c\x7e\x74\x72\x13\xc3\xba\xf4\x70\x5b\x73\x85\x97\x3f\xd7\x53\xd0\xfe\xf4\x58\x23\x18\x60\xc8\x31\xb7\xfd\xa1\x8a\xaf\x6d\xc3\x84\xe8\x41\xf4\x0f\xd6\x92\x93\x10\x0e\xa9\x5d\x8a\xec\x5a\x62\xee\x70\xf8\x5e\x44\x18\x68\x26\x3e\x7c\x6d\x6c\x98\x78\x46\x3d\x50\xe2\x03\x25\x3e\x50\xe2\x03\x25\xa2\xa1\x4b\xa9\xb3\xb1\x4b\xa9\xd3\xe3\xfa\x86\x68\x2d\xb8\x50\x35\x2e\x9f\xb8\x86\x6e\x83\x02\x59\xe8\xf5\x0a\x1b\xf8\x80\xb7\xfb\x90\xaa\x73\xe8\x1f\xa1\xd4\xbe\xe6\x03\x33\x84\x12\xe9\x4d\xb6\x96\xbd\x4a\xeb\xb6\xdd\x3e\xcf\xc4\x28\xda\x69\x44\x48\x62\x85\x73\xcf\x84\x5c\x64\x05\x03\xdf\x24\xcb\xe2\xc9\xc9\xd5\xd5\x17\xb9\x2d\xf9\xa6\x04\xfb\xf1\x5b\x97\x9d\x4a\x4f\x24\x55\xeb\xf0\x41\x34\x65\x98\x5c\xa6\xe2\x6a\x0f\x82\x9b\x40\x34\xa0\xaf\x3b\x23\xe5\xda\x60\x3e\x40\x31\x07\xe1\xcb\xcf\xc5\x8b\x77\xcf\x7f\x23\x89\x23\x45\xb6\x57\xd2\x48\x91\xdd\x47\x90\x5b\xda\x3d\xd4\x29\xf4\x96\x65\xcc\x24\x84\xaf\xab\x84\x99\x36\xf5\xae\xcb\x17\xab\xfd\xeb\x2c\x5d\xca\xe2\xe1\x46\x69\xdb\x92\x9c\x48\xdd\xbd\x4f\x50\x5f\x6e\xfe\xee\x77\x6e\x7a\x48\xe2\x87\x24\xfe\x53\x25\xf1\x6d\x0e\x00\xf5\x17\xe5\x1a\x50\xf9\x3b\xc7\x1c\xaf\x40\x79\x09\x3b\x99\xf8\x68\x30\xf9\x87\x6f\xc9\x0f\x75\x4f\x7e\xd0\x8f\x34\xa3\x4c\xd5\xaf\x9b\xd0\xaf\x13\xe1\xf2\x9b\x67\x66\xfb\xdb\xb0\xd7\x31\xc8\x7d\x37\x51\xb3\x8b\x77\x2c\xe5\x80\xaa\xde\x47\xfb\x03\x69\x6b\x9e\x71\x0c\xe8\xf4\xc3\xe5\x50\x5a\xa7\x4f\x19\x03\x20\xc2\xbf\x1d\x4a\x35\x11\xdc\x60\xca\x41\xb9\xef\xe5\x45\x9e\x82\xea\x6b\x1c\x50\x97\x29\x61\xd5\x0d\x3c\xe3\x68\x3d\x3c\x29\x5b\xbd\x96\xd1\xd7\x35\x0b\xdb\xa2\x0d\x70\xf3\x8b\x60\x45\x0e\x6f\x18\xa6\xf9\xbd\xbd\xb4\xd1\x12\x48\xf9\xd2\x86\x10\xd0\xfa\x9d\xc8\x76\x8f\x58\x2e\x00\x67\xff\x52\xd4\xc0\x0f\xd5\x09\x49\x41\xe5\x6a\x27\xa0\xe0\xbf\x05\xe8\xdd\x0b\x11\x6d\x84\xc2\x2b\x88\xd1\xec\xf9\x5b\x3a\x1b\xb7\xff\xc7\x2a\x5b\x5f\x57\xd9\x7a\x6f\x86\x77\xf1\xa9\x14\x93\x10\x17\x66\x2d\x14\xfd\x5f\x19\xa3\xe1\xe5\xf3\x92\x5e\x36\xf3\x14\x0c\x6e\x1e\x47\xb1\x42\x1b\x50\x17\x82\xc1\x1d\x62\xaf\xa2\xe9\xa8\xdc\x75\x24\x7d\xab\x44\x21\x3d\xce\x9e\xcd\xdc\x0a\xb4\x16\xc4\xf6\xf8\x6f\x21\xdc\xaf\xea\x3b\x6b\xa7\xad\xfd\x14\x20\x40\x9d\x2b\x90\xa0\x77\x59\x1e\x20\xef\xb0\x17\x20\xca\x57\x76\x97\xf7\x06\x54\xfc\x9a\x63\xf7\x25\x75\x03\x2a\xf5\x70\xf9\x47\x7e\x46\xf5\xee\xc7\x07\x6c\xc8\x7a\xd4\x4e\xb8\x32\xc0\xed\xfa\xe8\x51\x8b\x1b\x24\x91\x36\xd8\x14\x63\x10\xf7\xc7\xc3\x05\xbf\x70\xea\x7e\xbe\xf8\xde\x47\x55\x7b\x6f\x36\x36\xf3\x01\x62\xeb\x9f\x94\x5b\xb2\xb9\xc3\x10\xd3\x45\xfa\x1e\x88\x71\x51\x36\x98\x9b\x76\xc2\xf1\xec\x9a\xce\x2f\x5b\x34\x5c\xc0\xb2\xe4\x16\xb7\xb2\x13\x8e\x38\x42\xa8\x9f\x5e\xe3\xca\xaf\x23\x97\x7b\xa7\x53\xbb\xc9\x34\x44\x5a\x4d\xb5\x36\xa6\xde\xb6\x6c\x6f\x8c\x9e\x1f\xd7\x3f\x95\x30\x82\x08\x16\xa3\x9f\xde\x2c\x5c\x5b\x55\x17\x2e\x4a\x41\xf7\xda\x4e\x03\x03\x62\x84\xda\xd7\x12\x0d\xda\x7a\xe9\xf5\x72\x49\x39\x35\xdb\x18\x9d\x0b\x6e\x9d\x6b\xb6\xd2\xb2\xb5\xf3\xf8\x77\x8b\x3e\x67\x63\x29\x75\xd4\xf8\xf6\x5b\x90\x4c\x6c\xed\x71\xe3\xde\xdd\xdb\x76\x45\x6e\x73\xf6\xfb\x56\xf5\x3d\x06\x64\x08\x8a\x2d\xe7\x29\xc1\x3a\x46\x73\xeb\x15\xc8\x25\xc3\xc6\xdd\xd0\xfa\x86\xda\xbf\x6e\x8d\x3f\xae\x67\x48\x13\x42\xb5\x01\xe5\xbf\x5b\xc9\x76\x3e\x91\x67\x56\x8f\x2b\x5f\x5a\x05\xec\x88\x13\x83\x86\x8b\x03\x05\x4c\xe0\xcc\xab\x6a\x68\x5e\xed\xd1\xef\x69\x9e\xd3\x6c\xab\x05\x8f\xba\xd2\xf1\xe6\x38\x3c\x0d\x8f\x67\xdd\x51\x8b\x82\xb1\x85\x60\x94\x6c\x63\x34\xfb\x6e\x79\x2e\xcc\x42\x81\x06\x6e\x7c\x49\xac\xda\xa7\x07\x0b\x33\x08\x36\x65\xa5\x13\x64\x54\xbd\xf4\xde\xad\xf6\xc4\x3e\x40\xba\x16\xe2\x32\x28\x14\x7b\x69\xd3\x25\x8e\xa2\xf9\xc9\x3f\xc2\xe3\xf0\x38\x9c\x97\x6f\x4e\xa3\x20\xea\x59\xd4\xdb\x13\xaa\xbf\xb2\x28\xab\xff\x2a\xfd\xef\xac\x9f\x7b\xe8\x2a\x27\x56\x88\x1c\xd0\x96\x04\x42\xb9\x1d\xb7\xa8\xae\xf6\x47\xd0\x5b\x18\x38\xfb\x81\xb3\x6d\xe7\x9a\x63\x74\x91\x06\xd6\xc4\xca\x44\x3b\xc1\x78\x73\x12\xce\xcf\xc2\x93\x43\x2e\x84\x2b\xd2\x42\xa3\xb3\x34\x54\x60\xab\x50\x4b\xc2\xf6\xf0\xf8\x72\x7e\x96\xf5\xe4\x2b\x53\x43\x7b\x0c\xf3\x57\x2e\x6a\x3f\xb3\x9e\x56\x23\xb1\x59\xbf\x8c\x6c\x16\x0d\x2d\x78\x48\x04\xd7\x82\x41\xc8\x68\xaa\xb0\xa2\xa0\x2b\x45\x3b\x0d\x91\x93\x48\x1a\x89\xc9\x79\xea\x04\x1e\x9d\x67\x78\x38\x70\x9c\x32\x08\x18\x5d\x02\xd9\x12\xe6\xc7\x80\x47\xe6\xbb\x51\x4d\x46\xb6\x38\xda\x8f\x06\xca\x41\xeb\x85\x3d\xd9\xb4\xc7\xda\xc8\x7e\x0b\x26\xee\x04\x90\xfb\x76\x64\x23\x1c\x67\xdb\x6e\xe7\x90\x12\x84\x2c\xa1\x53\xcc\xbe\x05\x86\xb7\x3f\x02\x11\x3c\xd3\x31\x3a\x6d\xcb\xb8\x8b\x81\xb1\xee\x25\xa6\xac\x50\xf0\xd3\x5a\x81\x5e\x0b\x96\xc5\xe8\xb4\xd5\xaf\x8b\xf2\x64\xe0\xf5\xcf\xbd\x7e\x46\x37\x70\x4b\x33\xd7\x80\x99\x59\xff\x59\x0c\xfd\x22\x58\xa6\x9e\xc6\xa5\xd7\xf5\xf3\xf4\x92\xce\x9a\x99\x56\x9d\xb3\x59\xb3\x0f\x91\x42\x51\xb3\x7d\x23\xb8\x81\x2b\x6f\xc1\x96\xda\x95\x6d\x7f\x3f\x3b\x3b\x7d\xb6\xbb\x51\x29\xf8\xeb\x89\x9e\x73\xc1\x2f\x84\x30\x1d\x1e\x2c\xbb\x7e\xd6\xa0\xda\x63\x0c\xa8\x9c\xf2\xb2\x12\x7c\xab\x30\x81\x05\x28\x2a\x32\x6f\x09\xeb\x35\xac\x4c\x1d\xd8\xfd\xc6\x5c\x4a\xea\xff\x89\xa5\xbd\x12\xd7\xb1\xf1\xb5\x2e\x96\x43\xc7\xf7\xb6\x0e\x62\x9b\xc6\xf6\xf3\xdf\x03\x00\x00\xff\xff\x96\x36\xf6\x58\x76\x33\x00\x00"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/grafana"].(os.FileInfo),
		fs["/namespace.yaml"].(os.FileInfo),
		fs["/prometheus"].(os.FileInfo),
	}
	fs["/prometheus"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/prometheus/alertmanager.yaml"].(os.FileInfo),
		fs["/prometheus/kube-stats-metrics.yaml"].(os.FileInfo),
		fs["/prometheus/node-exporter.yaml"].(os.FileInfo),
		fs["/prometheus/pushgateway.yaml"].(os.FileInfo),
		fs["/prometheus/server.yaml"].(os.FileInfo),
	}

	return fs
}()

type vfsgen۰FS map[string]interface{}

func (fs vfsgen۰FS) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {
	case *vfsgen۰CompressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &vfsgen۰CompressedFile{
			vfsgen۰CompressedFileInfo: f,
			gr:                        gr,
		}, nil
	case *vfsgen۰FileInfo:
		return &vfsgen۰File{
			vfsgen۰FileInfo: f,
			Reader:          bytes.NewReader(f.content),
		}, nil
	case *vfsgen۰DirInfo:
		return &vfsgen۰Dir{
			vfsgen۰DirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}

// vfsgen۰CompressedFileInfo is a static definition of a gzip compressed file.
type vfsgen۰CompressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *vfsgen۰CompressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰CompressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰CompressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *vfsgen۰CompressedFileInfo) Name() string       { return f.name }
func (f *vfsgen۰CompressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *vfsgen۰CompressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰CompressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰CompressedFileInfo) IsDir() bool        { return false }
func (f *vfsgen۰CompressedFileInfo) Sys() interface{}   { return nil }

// vfsgen۰CompressedFile is an opened compressedFile instance.
type vfsgen۰CompressedFile struct {
	*vfsgen۰CompressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *vfsgen۰CompressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.CopyN(ioutil.Discard, f.gr, f.seekPos-f.grPos)
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *vfsgen۰CompressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.seekPos = 0 + offset
	case io.SeekCurrent:
		f.seekPos += offset
	case io.SeekEnd:
		f.seekPos = f.uncompressedSize + offset
	default:
		panic(fmt.Errorf("invalid whence value: %v", whence))
	}
	return f.seekPos, nil
}
func (f *vfsgen۰CompressedFile) Close() error {
	return f.gr.Close()
}

// vfsgen۰FileInfo is a static definition of an uncompressed file (because it's not worth gzip compressing).
type vfsgen۰FileInfo struct {
	name    string
	modTime time.Time
	content []byte
}

func (f *vfsgen۰FileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰FileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰FileInfo) NotWorthGzipCompressing() {}

func (f *vfsgen۰FileInfo) Name() string       { return f.name }
func (f *vfsgen۰FileInfo) Size() int64        { return int64(len(f.content)) }
func (f *vfsgen۰FileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰FileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰FileInfo) IsDir() bool        { return false }
func (f *vfsgen۰FileInfo) Sys() interface{}   { return nil }

// vfsgen۰File is an opened file instance.
type vfsgen۰File struct {
	*vfsgen۰FileInfo
	*bytes.Reader
}

func (f *vfsgen۰File) Close() error {
	return nil
}

// vfsgen۰DirInfo is a static definition of a directory.
type vfsgen۰DirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *vfsgen۰DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *vfsgen۰DirInfo) Close() error               { return nil }
func (d *vfsgen۰DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *vfsgen۰DirInfo) Name() string       { return d.name }
func (d *vfsgen۰DirInfo) Size() int64        { return 0 }
func (d *vfsgen۰DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *vfsgen۰DirInfo) ModTime() time.Time { return d.modTime }
func (d *vfsgen۰DirInfo) IsDir() bool        { return true }
func (d *vfsgen۰DirInfo) Sys() interface{}   { return nil }

// vfsgen۰Dir is an opened dir instance.
type vfsgen۰Dir struct {
	*vfsgen۰DirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *vfsgen۰Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *vfsgen۰Dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
