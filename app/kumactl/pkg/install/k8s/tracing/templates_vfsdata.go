// Code generated by vfsgen; DO NOT EDIT.

// +build !dev

package tracing

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
			modTime: time.Date(2020, 7, 25, 5, 23, 0, 311392340, time.UTC),
		},
		"/jaeger": &vfsgen۰DirInfo{
			name:    "jaeger",
			modTime: time.Date(2020, 7, 25, 5, 23, 0, 311252278, time.UTC),
		},
		"/jaeger/all-in-one-template.yaml": &vfsgen۰CompressedFileInfo{
			name:             "all-in-one-template.yaml",
			modTime:          time.Date(2020, 7, 25, 5, 23, 0, 311304623, time.UTC),
			uncompressedSize: 4945,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd4\x56\x4d\x73\xdb\x36\x10\xbd\xf3\x57\xec\x48\x97\x76\xc6\x24\x4d\x35\xfe\x08\x7b\x52\x64\xb7\x51\xe3\x91\x34\x96\xd2\x4c\x7a\xc9\x40\xd0\x8a\x44\x03\x02\x0c\xb0\x94\xa2\x66\xf2\xdf\x3b\xa4\xc8\x88\x94\xa8\xc6\x76\x9c\xd6\xe5\x8d\xc0\xdb\xdd\xb7\x8b\xb7\x0b\x74\xe1\x05\xb3\xb8\x00\xad\x20\x26\x4a\x6d\xe8\xfb\x91\xa0\x38\x9b\x7b\x5c\x27\xfe\x9f\x0c\x23\x34\x64\x18\x17\x2a\x2a\xff\xdc\xf7\xd9\x1c\x8d\x42\x42\xeb\xcf\xa5\x9e\xfb\x09\xb3\x84\xc6\x67\x52\xba\x42\xb9\x5a\x61\x05\xdc\xad\xb8\x84\x49\x2a\x19\xa1\xb7\x49\xa4\xe3\x74\x9d\x2e\x0c\x74\xba\x31\x22\x8a\x09\x7a\xa7\xc1\x85\xdb\x3b\x0d\x9e\xc3\x2c\x46\xf8\xad\xb0\x85\x7e\x46\xb1\x36\xb6\x80\xde\x08\x8e\x2a\x27\x99\xa9\x05\x1a\xa0\x18\xa1\x9f\x32\x1e\x63\xb5\x73\x02\xbf\xa3\xb1\x42\x2b\xe8\x79\xa7\xf0\x43\x0e\xe8\x94\x5b\x9d\x1f\x7f\x86\x8d\xce\x20\x61\x1b\x50\x9a\x20\xb3\x08\x14\x0b\x0b\x4b\x21\x11\xf0\x23\xc7\x94\x9c\x2e\x08\x05\x5c\x27\xa9\x14\x4c\x71\x84\xb5\xa0\xb8\x08\x53\x3a\xf1\xe0\x6d\xe9\x42\xcf\x89\x09\x05\x0c\xb8\x4e\x37\xa0\x97\x75\x14\x30\x2a\xe8\xe6\x75\x0c\x7d\x7f\xbd\x5e\x7b\xac\xa0\xe9\x69\x13\xf9\x72\x0b\xb2\xfe\xcd\x70\x70\x3d\x9a\x5e\xbb\x3d\xef\xb4\x80\xbf\x56\x12\xad\x05\x83\x1f\x32\x61\x70\x01\xf3\x0d\xb0\x34\x95\x82\xb3\xb9\x44\x90\x6c\x0d\xda\x00\x8b\x0c\xe2\x02\x48\xe7\x4c\xd7\x46\x90\x50\xd1\x09\x58\xbd\xa4\x35\x33\x08\x0b\x61\xc9\x88\x79\x46\x8d\x22\x95\xbc\xf2\xf4\x6c\x03\xa2\x15\x30\x05\x9d\xfe\x14\x86\xd3\x0e\xbc\xe8\x4f\x87\xd3\x13\x78\x33\x9c\xbd\x1c\xbf\x9e\xc1\x9b\xfe\xed\x6d\x7f\x34\x1b\x5e\x4f\x61\x7c\x0b\x83\xf1\xe8\x6a\x38\x1b\x8e\x47\x53\x18\xff\x02\xfd\xd1\x5b\x78\x35\x1c\x5d\x9d\x00\x0a\x8a\xd1\x00\x7e\x4c\x0d\x5a\xeb\x74\x73\x92\x22\x2f\x20\x2e\x3c\x98\x22\x36\x0a\xb3\xd4\x5b\x42\x36\x45\x2e\x96\x82\x83\x64\x2a\xca\x58\x84\x10\xe9\x15\x1a\x25\x54\x04\x29\x9a\x44\xd8\xfc\x10\x2d\x30\xb5\x00\x29\x12\x41\x8c\x8a\xff\x22\x25\xa7\xdb\x38\x12\xa7\xeb\x38\x2c\x15\xe5\xc1\x87\xb0\x0a\x9c\xf7\x42\x2d\x42\xb8\x11\x96\x1c\x41\x98\xd8\xd0\x01\x70\xa1\x0e\x62\x69\x6a\xfd\x55\xe0\x00\x00\x6c\xd1\x57\x98\x4a\xbd\x49\x50\x51\xb1\x98\x20\xb1\x05\x23\x16\x16\x7f\x00\x8a\x25\x18\xc2\x56\xcf\xb5\x25\x9b\x32\x8e\x21\x7c\xfa\x04\xde\xa8\xfa\x85\xcf\x9f\x4b\x84\x64\x73\x94\xb6\x72\x01\x79\xd4\x3d\x1f\xc5\x9a\xb7\x6b\x24\x4f\x68\xbf\x25\x54\x1b\x2c\x97\xa9\x56\xa8\x28\x84\x5d\x7f\x15\xf8\xbc\xba\x55\x50\x83\x85\x82\x6c\x08\x41\xb9\x62\x51\x22\x27\x6d\x76\xb4\x12\x46\x3c\xbe\xd9\xe3\xda\xc6\xd6\x92\x61\x84\xd1\x66\x87\xa2\x4d\x8a\x21\xdc\x22\x37\xc8\x08\xcb\xe5\xaa\xc7\x6b\x11\xf6\xaa\xd9\x56\x9c\x63\x05\xba\x47\x91\xee\x5b\xa8\xd2\x42\x29\x5d\x0a\xac\xc9\x26\x35\x3a\x41\x8a\x31\x2b\xdc\x58\x6e\x58\x9e\x6c\x87\x4c\x86\x9d\x7f\x00\xa6\xda\x50\x08\x9d\xe0\xfc\xfc\xf2\x7c\x87\xab\x1f\x4a\xfe\x71\xad\xf2\x09\x82\x66\x2f\xa8\x0b\x00\xa8\x56\xcd\xc5\x6a\x6b\x9b\xf5\x60\x7c\x73\x73\x3d\x98\x8d\x6f\xdf\xfd\x31\x9c\xbc\x1a\x8e\xde\xbd\x9c\xcd\x26\xef\x26\xe3\xdb\x59\x8b\x11\xc0\x8a\xc9\x2c\xe7\xfd\xfc\x59\x10\x74\x0e\x10\x22\x61\xd1\x97\x4a\x56\x03\x7e\x57\xa7\x30\xf0\x82\xcb\x03\xa3\xa3\xd5\x2f\xaa\xa1\x0d\xd9\x76\xfe\x5f\x92\x9e\x14\x35\x3a\xbb\xb8\x38\x6b\xa5\x9c\x1a\x4d\x9a\x6b\x19\xc2\xeb\xab\xc9\x5d\x3c\x9d\x5f\xfe\x14\x3c\x9a\xa7\xde\xe3\x78\x3a\xbb\xb8\x38\xac\x5c\xd3\xd3\x6c\x70\x27\x4f\x85\x94\x1e\xc7\x55\x2e\x82\x07\x78\x32\xc8\x16\x42\xa1\xb5\x13\xa3\xe7\xd8\x76\xb8\xf9\x65\xf7\x2b\x52\xdb\x16\x40\xca\x28\x0e\xa1\xe3\x1f\xca\x0f\x4a\xc1\x84\x10\x3c\xeb\x9d\x3f\x6f\xd9\x17\x4a\x90\x60\xf2\x0a\x25\xdb\x4c\x91\x6b\xb5\xb0\x21\x9c\x1d\x4c\xf4\xc6\x30\x9f\xa2\x59\x09\x8e\x5f\x9d\xe4\xee\x87\x0c\xcd\xe6\xc9\xcd\xf3\x1d\xab\xfa\xd4\x78\xc8\xe0\x6e\x24\xb8\xd7\x98\xd5\x38\x29\x20\x6e\x7e\x7e\xce\xfe\x99\x5c\x9e\xd6\x97\x8e\x28\x84\x98\x89\x90\x0e\x95\x7a\xc8\xf7\x3b\xdc\x75\xd5\x15\x34\x90\x59\xfe\xf8\x1c\x4e\x1e\x49\x18\x5c\xcb\x2d\xfb\x27\x27\x8e\x26\xb3\x6f\x16\xc8\x7e\xa2\x47\x44\xb2\x8f\x76\x89\xc7\x4c\x29\x94\x07\x9a\xc9\xfb\xf8\xe2\xbe\xb2\x69\xd8\x1c\x8d\xd9\xaa\xd1\xdc\xf6\xf2\x01\xf1\x2e\xbf\x1e\xef\x2f\x91\xbe\x17\xea\x20\xe2\xde\x0c\xbd\x4b\xc0\x9a\xc9\xff\xba\x2d\x58\x54\x3d\x90\x9f\x52\x4b\xec\x58\x7d\x73\x3b\xd4\x13\x3c\xd2\x0a\x05\xa4\xd4\x86\x4b\xb1\x11\x4b\x3a\x90\xc8\xde\xcb\xe6\xd8\xdb\xa1\x2e\x91\x86\x49\x33\x56\x9e\x2a\xe3\x87\x51\xf6\x5e\x3d\x77\x89\xd2\x30\x69\x46\x99\x0b\xc5\xbe\xdc\x15\x8d\x20\xbd\xfb\x07\xe9\x1d\x4d\x45\x2d\x45\x64\xdb\x0a\x76\xdf\x26\xae\x99\xf0\x4a\xe6\x21\x8c\x76\x2d\xf0\x2f\x74\xda\x37\xf5\x54\x63\xbc\x3c\x9d\x6e\xaa\xd1\x7a\x78\x3b\x35\x72\xbb\xeb\x9d\xf2\x5d\xe6\xed\x7f\xa4\x0d\xe7\xef\x00\x00\x00\xff\xff\xbc\xd3\x57\xc6\x51\x13\x00\x00"),
		},
		"/namespace.yaml": &vfsgen۰FileInfo{
			name:    "namespace.yaml",
			modTime: time.Date(2020, 7, 25, 5, 23, 0, 311447403, time.UTC),
			content: []byte("\x0a\x2d\x2d\x2d\x0a\x61\x70\x69\x56\x65\x72\x73\x69\x6f\x6e\x3a\x20\x76\x31\x0a\x6b\x69\x6e\x64\x3a\x20\x4e\x61\x6d\x65\x73\x70\x61\x63\x65\x0a\x6d\x65\x74\x61\x64\x61\x74\x61\x3a\x0a\x20\x20\x6e\x61\x6d\x65\x3a\x20\x7b\x7b\x20\x2e\x4e\x61\x6d\x65\x73\x70\x61\x63\x65\x20\x7d\x7d\x0a"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/jaeger"].(os.FileInfo),
		fs["/namespace.yaml"].(os.FileInfo),
	}
	fs["/jaeger"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/jaeger/all-in-one-template.yaml"].(os.FileInfo),
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
