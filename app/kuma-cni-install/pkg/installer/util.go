package installer

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/facebookgo/atomicfile"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func CreateFileWatcher(dir string) (watcher *fsnotify.Watcher, fileModified chan bool, errChan chan error, err error) {
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}

	fileModified, errChan = make(chan bool), make(chan error)
	go watchFiles(watcher, fileModified, errChan)

	if err = watcher.Add(dir); err != nil {
		if closeErr := watcher.Close(); closeErr != nil {
			err = errors.Wrap(err, closeErr.Error())
		}

		return nil, nil, nil, err
	}

	return
}

func watchFiles(watcher *fsnotify.Watcher, fileModified chan bool, errChan chan error) {
	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}

			fileModified <- true
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			errChan <- err
		}
	}
}

// WaitForFileMod is a function which waits until a file is modified (returns nil),
// the context is cancelled (returns context error), or returns error
func WaitForFileMod(ctx context.Context, fileModified chan bool, errChan chan error) error {
	select {
	case <-fileModified:
		return nil
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReadCNIConfigMap is a function which will read CNI config from file
// and return the unmarshalled JSON as a map
func ReadCNIConfigMap(path string) (map[string]interface{}, error) {
	cniConfig, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cniConfigMap map[string]interface{}
	if err := json.Unmarshal(cniConfig, &cniConfigMap); err != nil {
		return nil, errors.Wrap(err, path)
	}

	return cniConfigMap, nil
}

// GetPlugins is a function which given an unmarshalled CNI config JSON map,
// return the plugin list asserted as a []interface{}
func GetPlugins(cniConfigMap map[string]interface{}) ([]interface{}, error) {
	plugins, ok := cniConfigMap["plugins"].([]interface{})
	if !ok {
		return nil, errors.New("error reading plugin list from CNI config")
	}

	return plugins, nil
}

// GetPlugin is a function which given the raw plugin interface, returns
// the plugin asserted as a map[string]interface{}
func GetPlugin(rawPlugin interface{}) (map[string]interface{}, error) {
	plugin, ok := rawPlugin.(map[string]interface{})
	if !ok {
		return nil, errors.New("error reading plugin from CNI config plugin list")
	}

	return plugin, nil
}

// MarshalCNIConfig is a function which marshal the CNI config map and append
// a new line
func MarshalCNIConfig(cniConfigMap map[string]interface{}) ([]byte, error) {
	cniConfig, err := json.MarshalIndent(cniConfigMap, "", "  ")
	if err != nil {
		return nil, err
	}

	return append(cniConfig, "\n"...), nil
}

// Write atomically by writing to a temporary file in the same directory then renaming
func fileAtomicWrite(path string, data []byte, mode os.FileMode) error {
	f, err := atomicfile.New(path, mode)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return f.Close()
}

// Copies file by reading the file then writing atomically into the target directory
func fileAtomicCopy(srcFilepath, targetDir, targetFilename string) error {
	info, err := os.Stat(srcFilepath)
	if err != nil {
		return err
	}

	input, err := ioutil.ReadFile(srcFilepath)
	if err != nil {
		return err
	}

	return fileAtomicWrite(filepath.Join(targetDir, targetFilename), input, info.Mode())
}
