// Code generated by vfsgen; DO NOT EDIT.

// +build !dev

package migrations

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

// Migrations statically implements the virtual filesystem provided to vfsgen.
var Migrations = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2021, 1, 12, 10, 8, 10, 94271006, time.UTC),
		},
		"/1579518998_create_resources.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1579518998_create_resources.up.sql",
			modTime:          time.Date(2020, 3, 21, 15, 26, 20, 56719749, time.UTC),
			uncompressedSize: 299,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x8c\x8e\x4f\x0b\x82\x40\x10\x47\xef\x7e\x8a\xdf\x51\x61\x0f\x76\xee\x64\xb1\x81\x64\x16\xba\x41\x1e\x97\x65\x48\x0f\xea\xb2\xb3\x49\x7d\xfb\x68\xfb\x07\x1d\xc2\x39\x0e\x6f\xde\xbc\x75\x25\x33\x25\xa1\xb2\x55\x21\x91\x6f\x50\xee\x15\xe4\x29\xaf\x55\x0d\x47\x3c\x5e\x9c\x21\x46\x1c\x01\xc0\xa0\x7b\xc2\x6b\x26\xed\x4c\xab\x5d\xbc\x48\xd3\x24\xdc\x94\xc7\xa2\x10\x1f\x8c\xad\x36\xf4\x1f\xeb\x89\xdb\x19\x36\x7f\xb3\x73\x9e\x4e\xe4\xb8\x1b\x87\x80\x75\x83\xa7\x33\xb9\x1f\x82\x2d\x99\xb7\xc8\xd3\xd5\x3f\xb7\x87\x2a\xdf\x65\x55\x83\xad\x6c\x10\x3f\xca\xc5\xb7\x5f\x84\x46\x11\x12\x92\x28\x59\xde\x03\x00\x00\xff\xff\x88\x1c\x8d\x52\x2b\x01\x00\x00"),
		},
		"/1580128050_add_creation_modification_time.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1580128050_add_creation_modification_time.up.sql",
			modTime:          time.Date(2020, 3, 21, 15, 26, 20, 56811003, time.UTC),
			uncompressedSize: 165,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x28\x4a\x2d\xce\x2f\x2d\x4a\x4e\x2d\x56\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x53\x48\x2e\x4a\x4d\x2c\xc9\xcc\xcf\x8b\x2f\xc9\xcc\x4d\x55\x08\xf1\xf4\x75\x0d\x0e\x71\xf4\x0d\x50\xf0\xf3\x0f\x51\xf0\x0b\xf5\xf1\x51\x70\x71\x75\x73\x0c\xf5\x09\x51\xc8\xcb\x2f\xd7\xd0\xb4\xe6\x22\x68\x60\x6e\x7e\x4a\x66\x5a\x66\x32\x29\x86\x02\x02\x00\x00\xff\xff\x56\x69\x01\xb8\xa5\x00\x00\x00"),
		},
		"/1589041445_add_unique_id_and_owner.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1589041445_add_unique_id_and_owner.up.sql",
			modTime:          time.Date(2020, 5, 13, 13, 58, 0, 972458693, time.UTC),
			uncompressedSize: 973,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x9c\x91\x51\x4f\xdb\x30\x14\x85\xdf\xfd\x2b\xce\x9e\x4a\xa5\x74\xa2\xbc\x56\x3c\x78\xb1\xbb\x45\x4b\x9c\xca\x49\x35\xf1\x84\xac\x70\x69\x2a\x42\x12\xd9\x61\x8c\x7f\x3f\xc5\x09\xab\x41\x8c\x89\xbd\xde\xeb\x73\xee\x77\x8e\x79\x5a\x4a\x8d\x92\x7f\x49\x25\x2c\xb9\xee\xc1\x56\xe4\x18\x00\x08\x9d\xef\x10\xe7\xaa\x28\x35\x4f\x54\x79\xda\x5e\xf7\x77\xf4\x14\xf9\x37\x5c\x88\xbf\x3f\xc1\x4e\x27\x19\xd7\x57\xf8\x2e\xaf\x70\xd6\x9a\x7b\x8a\x70\x4f\xae\x8e\x30\x3c\xf5\xb4\xdc\xb0\x7f\xde\x4e\xf7\x99\xc2\x28\x74\xbd\xa9\xe8\x3d\xc1\x04\xe2\xdf\x77\x8f\x2d\xd9\xeb\x51\x85\x9f\xc6\x56\xb5\xb1\x67\xeb\xf3\xf3\xe5\x87\xd4\x23\xe6\xff\xab\xc7\x78\x1f\x54\xff\xa9\x70\x72\xb8\xbd\xc3\x36\xd7\x32\xf9\xaa\xa6\xf2\x4e\x99\xa2\x80\x30\x0a\xee\x2d\xa1\xe5\x56\x6a\xa9\x62\x59\x9c\x2e\xbc\x51\x3b\x72\x05\x21\x53\x59\x4a\xc4\xbc\x88\xb9\x90\x1b\xc6\xa6\x01\xdb\xea\x3c\x0b\xf0\x7e\x7c\x93\x5a\x7a\x15\x2e\xb1\x10\x66\x30\x7d\x63\x5a\x4a\x5a\x77\x3c\xd4\xc3\x62\xc3\xd8\x6a\x05\x47\xc3\x84\x81\xdb\xce\xc2\x34\x4d\x70\x9d\x7e\x55\xd4\x0f\xc8\xc8\xd5\x6c\xbf\x13\xbc\x0c\xc2\xc3\xae\x59\x21\xcb\xf0\xbb\x2e\x61\x2f\x3e\x7b\x60\xdf\x4c\xf0\x15\x7e\xe3\x43\x04\x9b\x67\xb0\xd1\x7e\xf1\x8a\x1d\xf6\x62\xc6\x9f\x2d\x47\x8b\xb5\xb7\x60\x00\x57\x62\x9c\xbf\x34\x98\xc7\xeb\x69\xfc\xe9\x79\xee\x53\xae\x70\x43\x0d\x0d\x84\x1b\xd3\x1e\x9a\x63\x7b\x08\x2b\xee\x5a\x72\x78\x3c\x0e\x75\xf7\x30\x57\xb1\x7c\xb7\xd0\x20\x71\x52\x40\xed\xd3\x74\xbe\x1d\x04\x7e\x6b\xe1\xb9\x5e\x2e\x5e\xa1\xfe\x0e\x00\x00\xff\xff\x17\x5c\x38\xd4\xcd\x03\x00\x00"),
		},
		"/1592232449_add_leader_table.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1592232449_add_leader_table.up.sql",
			modTime:          time.Date(2020, 7, 22, 13, 30, 56, 565733757, time.UTC),
			uncompressedSize: 217,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x74\x8e\x31\xab\xc2\x30\x1c\xc4\xf7\x7c\x8a\x1b\x5b\x78\xbc\xe1\x41\xa7\x4e\x49\xde\x9f\x1a\xd4\xa8\x31\x2a\x99\x4a\x6d\x33\x88\x36\x81\x54\xeb\xd7\x17\x6c\x47\x1d\xef\xf8\x71\xf7\x93\x86\xb8\x25\x58\x2e\x56\x84\x5b\x6c\xaf\x03\x32\x06\x00\xa1\xe9\x3d\xe4\x82\x1b\x2e\x2d\x19\x1c\xb9\x71\x4a\x57\xd9\x5f\x51\xe4\xd8\x1a\xb5\xe6\xc6\x61\x49\xee\xe7\x0d\x27\xdf\xc6\xd4\xd5\xa3\x4f\xc3\x25\x86\x3a\x3c\xfa\xb3\x4f\x10\xaa\x52\xda\x4e\x44\xd7\xdc\x1b\x08\x67\x89\x4f\x39\x3e\x83\x4f\x5f\xf6\x59\x5e\x32\x36\x8b\xed\x69\x77\x20\x2d\x67\xb7\x3a\x8d\x01\x9b\x93\xa6\x7f\x08\x37\x55\xbf\x1f\xbf\x4b\xf6\x0a\x00\x00\xff\xff\x9b\x4e\x1b\xbc\xd9\x00\x00\x00"),
		},
		"/1604501001_adjust_global_resources.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1604501001_adjust_global_resources.up.sql",
			modTime:          time.Date(2020, 11, 25, 7, 36, 58, 843884862, time.UTC),
			uncompressedSize: 565,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x74\x91\x41\x6f\xab\x30\x10\x84\xef\xfc\x8a\xb9\x91\x48\xe1\xfd\x81\xe8\x1d\x78\xb0\xc9\x8b\x9a\x42\x64\x5c\x55\xed\xa5\x72\x61\x13\x50\xc1\x46\xd8\x24\xca\xbf\xaf\x28\x6d\x43\xdb\xf4\xe8\xd1\xee\x37\x33\xeb\x20\xc0\xba\x36\xcf\xaa\x46\xc7\xd6\xf4\x5d\xce\x16\xb3\x5b\xb6\x25\x94\x2e\xf0\x68\x34\x63\x6f\x3a\x68\x73\x9a\x43\x33\x17\x16\xce\xa0\x54\x47\x86\xd2\xe0\xa6\x75\x67\x1c\x55\xdd\x33\x2a\x8d\x66\x58\xcb\x4d\xdd\x37\xfa\x0f\x64\x59\x59\x34\xd5\xa1\x53\xae\x32\x1a\x7d\x5b\x28\xc7\x16\xae\x34\x96\x2f\x5e\x5e\xb8\x95\x24\x20\xc3\x7f\x5b\x9a\x24\x88\x45\xba\x43\x94\x26\x99\x14\xe1\x26\x91\x30\x27\xcd\xdd\xd3\xfe\x65\x89\x20\x40\xd1\x99\x16\xb9\xd1\xd6\x75\xaa\xd2\x0e\xd6\xe0\xc4\xc8\xd5\x87\xc9\x85\xb3\xf4\xee\x76\x71\x28\xa7\xe4\x8c\xe4\x98\xf3\x2f\x7c\x1f\xf7\xff\x49\x10\xdc\xb9\xe5\xe1\x3d\xd4\xf6\x91\x8a\x4f\x61\xa8\xef\xff\x02\x19\x23\x7d\x47\x8d\xea\x0f\xe0\x57\xf9\x1d\x7b\xbd\x7a\x18\xc7\xd7\x9a\x63\x95\x0a\xda\xac\x13\xdc\xd0\x03\x66\xa3\xaa\x55\xc3\x8b\x49\x90\xc5\xc4\x67\x0e\x41\x2b\x12\x94\x44\x94\x4d\xbf\x76\xdc\x19\xa7\xc7\xb9\x34\x41\x4c\x5b\x92\x84\x28\xcc\xa2\x30\xa6\xb7\x23\xab\xb6\xad\xcf\x70\x25\x4f\x2f\xad\x0e\xaa\xd2\x4b\xcf\x7b\x0d\x00\x00\xff\xff\xfc\xc8\xa8\xa1\x35\x02\x00\x00"),
		},
		"/1605012491_add_trigger.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1605012491_add_trigger.up.sql",
			modTime:          time.Date(2021, 1, 12, 10, 5, 24, 497136855, time.UTC),
			uncompressedSize: 935,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x7c\x93\xc1\x6e\xa3\x3c\x14\x85\xf7\x7e\x8a\xb3\x88\x44\x22\x91\x3e\x40\x51\xfa\x8b\xc2\x0d\xcd\x2f\x6a\x22\x07\xd4\xd9\x21\x02\x6e\x4a\x07\xd9\x19\xec\xb4\x9d\xb7\x1f\x61\x92\x0e\x69\xa5\x61\x85\xed\x73\x8f\x7d\x3f\x9d\x1b\x09\x0a\x73\x42\x26\x20\x68\x9b\x86\x11\x61\x5d\xf0\x28\xdf\x64\x1c\x4a\xdb\xf6\xf9\x77\x29\xdf\xa4\xb2\xf3\x05\x04\xe5\x85\xe0\x3b\xe4\x62\x93\x24\x24\x10\xee\x30\x9b\x31\x16\x53\x94\x86\x82\x18\x00\x34\x95\xad\xf0\x6a\xb4\x0a\xdc\xd2\x19\xb4\x75\x65\x5b\xad\xce\xdb\xec\x9e\x92\x0d\x67\xee\x78\xb9\x44\xa4\xd5\x9b\xec\x2d\xec\x8b\x84\xee\x1a\xe8\x1e\x4a\xbe\xa3\xd7\xef\xb0\x1a\xff\xef\x32\xee\x63\x5f\x19\xd9\x40\x2b\x27\xfa\xd9\xaa\x06\xfa\x19\x55\x3d\x98\xde\x5c\x7c\x42\xb7\xc4\x0a\x31\xa5\x94\xd3\x7f\x98\x7e\xcb\x3b\x64\x69\x3c\x98\x7e\x93\x6f\xf8\x8e\x44\x3e\x5c\x5b\x6c\xe3\x70\x2c\x5c\xde\x81\xd3\xd3\xa7\x7c\xb3\xc6\x3c\x4f\xca\x6c\x8b\x15\xbc\xd1\xde\x5b\x20\x7f\x20\xce\x2e\x17\xb8\xb6\x57\x43\x45\x69\x75\x39\x34\x3a\xcf\xd2\x78\x31\x42\xa0\x74\x47\xff\x52\x72\x7a\xba\x28\x79\x8c\xcd\x3a\x98\xc2\xb1\xfd\xa9\x1e\xe9\x5c\xb1\xac\x0c\x2a\x47\x07\xc6\xf6\xad\x3a\xdc\x7c\xc7\xbd\x72\xc0\xcb\xfd\xa9\xed\x9a\x52\xef\x5f\x65\x6d\xe7\x6c\x4a\xc5\xb3\xd5\xbe\x93\x9e\x9f\x27\x65\x1e\xde\xa7\x54\xf2\xf0\x91\xfc\x6b\xc9\x48\xd9\xf3\xe1\x00\x7c\x39\x1c\x7a\xf1\x7c\xd7\xd2\x22\x60\x9f\xaf\xa6\x0f\x59\x9f\xac\xc4\xf1\x50\x8e\x01\x9a\xd7\x2f\x95\x52\xb2\xf3\xaf\x1e\xb8\x70\xfa\x2d\x89\x75\x26\x1e\x27\x62\xaf\x97\x46\x9f\xfa\x5a\x8e\xb9\x33\x9e\x3f\xad\xba\xbd\xb5\xf2\xc3\x2e\xfe\x32\x12\xd2\x9c\x3a\x8b\xd6\xa0\x3d\x28\xdd\xcb\x06\xa6\x55\xb5\x84\x7d\x19\xb6\x0c\x2a\x85\x70\x9d\x93\x80\xed\xdb\xc3\x41\xf6\xae\x6e\x4c\x32\x78\x91\xa6\x01\x23\x1e\x07\x8c\xcd\x66\x48\x43\x9e\x14\x61\x42\x38\x76\xc7\x83\xf9\xd5\x05\x8c\x9d\x87\xe3\x92\xf8\x63\xaf\x9b\x53\x6d\x4d\x39\x9d\x0c\xe7\x38\xde\x71\x4e\x53\x26\xce\x69\x1a\xfe\xc6\xc4\x20\xe3\xb8\x34\x66\x5c\xc5\x3a\x13\xa0\x30\x7a\x80\xc8\x9e\x40\x3f\x28\x2a\x72\xc2\x56\x64\x11\xc5\x85\xa0\x2f\xb3\x17\xb0\x3f\x01\x00\x00\xff\xff\x0f\x3c\x0c\xea\xa7\x03\x00\x00"),
		},
		"/1610445956_update_notify_event.up.sql": &vfsgen۰CompressedFileInfo{
			name:             "1610445956_update_notify_event.up.sql",
			modTime:          time.Date(2021, 1, 12, 10, 8, 10, 94203823, time.UTC),
			uncompressedSize: 958,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\x94\x53\x5d\x6f\x9b\x30\x14\x7d\xf7\xaf\x38\x0f\x91\x08\x12\xe9\x0f\x28\x6a\x27\x16\x6e\x18\x13\x33\x91\x43\xd4\xc7\x88\x82\x4b\xe8\xa8\x9d\x61\xa7\x1f\xff\x7e\xc2\x34\x1d\x69\x26\x4d\xe3\xe9\xe2\x7b\xee\xb9\xc7\xe7\x5e\x2f\x05\x45\x05\x21\x17\x10\xb4\xce\xa2\x25\x61\xb5\xe5\xcb\x22\xcd\x39\x94\xb6\xed\xc3\xdb\x4e\x3e\x4b\x65\xe7\x3e\x04\x15\x5b\xc1\x37\x28\x44\x9a\x24\x24\x10\x6d\x30\x9b\x31\x16\xd3\x32\x8b\x04\x31\x00\xa8\x4b\x5b\xe2\xd1\x68\x15\xba\x5f\x47\xd0\x56\xa5\x6d\xb5\x7a\x3f\x66\x5f\x29\x49\x39\x73\xe9\xc5\x02\x4b\xad\x9e\x65\x6f\x61\xf7\x12\xba\xab\xa1\x7b\x28\xf9\x82\x5e\xbf\xc0\x6a\x7c\xdf\xe4\x3c\xc0\x7d\x69\x64\x0d\xad\x1c\xe8\x67\xab\x6a\xe8\x07\x94\xd5\x40\x7a\x75\xe2\x89\xdc\x2f\x6e\x10\x53\x46\x05\x7d\xc1\xf4\x5b\xdc\x22\xcf\xe2\x81\xf4\x02\x9e\xf2\x0d\x89\x62\x68\xbb\x5d\xc7\xd1\x58\xb8\xb8\x05\xa7\xbb\x0f\x78\xba\xc2\xbc\x48\x76\xf9\x1a\x37\xf0\x46\x7a\xcf\x47\xf1\x8d\x38\x3b\x35\x70\xd7\xbe\x71\x37\xdc\xdd\x1f\xdb\xae\xde\xe9\xfb\x47\x59\xd9\x39\x9b\xca\xf0\x54\xf9\x24\xbd\x60\xd0\x72\x35\x84\xc1\x79\xf6\x49\x9a\xfd\x7b\x76\x08\x3f\x65\xed\xdb\xe1\x54\x3b\x84\xfe\x68\x30\x65\x1b\xfa\x4f\x15\x53\x25\x9c\xee\xfe\xa2\x64\xaa\x66\x40\x5c\xaa\x99\x2a\x1a\x10\x53\x45\x3c\x46\xba\x0a\xa7\x03\x36\xb6\x3f\x56\xe3\x88\xcf\x16\xa2\x34\x28\xdd\x88\x61\x6c\xdf\xaa\xe6\xea\x72\x67\xfe\xed\xe9\xb8\x07\x5e\x00\x37\xa2\x4f\xa6\x0d\x8e\x78\x81\x33\xc6\x0f\xd9\x87\x26\x7a\x95\xd5\xd1\x4a\x1c\x9a\xdd\xb8\xe2\xf3\x6a\x5f\x2a\x25\xbb\xe0\xac\xbb\xef\xf0\x6b\x12\xab\x5c\xfc\x98\x80\xbd\x5e\x1a\x7d\xec\x2b\x39\xbe\x0c\xe3\x05\xd3\xaa\xeb\x6b\x2b\x5f\xad\xff\xc7\x01\x21\xcd\xb1\xb3\x68\x0d\xda\x46\xe9\x5e\xd6\x30\xad\xaa\x24\xec\x7e\x38\x32\x28\x15\xa2\x55\x41\x02\xb6\x6f\x9b\x46\xf6\xae\x6e\x7c\x6b\xe0\xdb\x2c\x0b\x19\xf1\x38\x64\x6c\x36\x43\x16\xf1\x64\x1b\x25\x84\x43\x77\x68\xcc\xaf\x2e\x64\xbf\x03\x00\x00\xff\xff\x42\xea\x50\x78\xbe\x03\x00\x00"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/1579518998_create_resources.up.sql"].(os.FileInfo),
		fs["/1580128050_add_creation_modification_time.up.sql"].(os.FileInfo),
		fs["/1589041445_add_unique_id_and_owner.up.sql"].(os.FileInfo),
		fs["/1592232449_add_leader_table.up.sql"].(os.FileInfo),
		fs["/1604501001_adjust_global_resources.up.sql"].(os.FileInfo),
		fs["/1605012491_add_trigger.up.sql"].(os.FileInfo),
		fs["/1610445956_update_notify_event.up.sql"].(os.FileInfo),
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
