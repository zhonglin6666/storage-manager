package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// IsFileExists - whether the file exist
// @filename: file full path name
func IsFileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return os.IsExist(err)
	}

	return !fi.IsDir()
}

// RunShellCommand exec shell command and output string format.
func RunShellCommand(shell string) (string, error) {
	glog.V(4).Infof("Run shell %s.", shell)

	cmd := exec.Command("/bin/sh", "-c", shell)
	buf, err := cmd.Output()
	output := string(buf)
	output = strings.TrimSuffix(output, "\n")

	if err != nil {
		glog.V(4).Infof("Run shell %s failed with error %s.", shell, err)
	}
	return output, err
}

func truncateID(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}

// GenerateRandomID returns an unique id
func GenerateRandomID() string {
	for {
		id := make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, id); err != nil {
			panic(err) // This shouldn't happen
		}
		value := hex.EncodeToString(id)
		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numberic and causes issues when
		// used as a hostname. ref #3869
		if _, err := strconv.ParseInt(truncateID(value), 10, 32); err == nil {
			continue
		}
		return value
	}
}

// CheckMapString check whether the map key's value type is string
func CheckMapString(m map[string]interface{}, key string) (string, error) {
	if _, ok := m[key]; !ok {
		return "", fmt.Errorf(key + " is nil")
	}
	if v, ok := m[key].(string); ok {
		return v, nil
	}
	return "", fmt.Errorf(key + " is not string type")
}

// CheckMapInt32 check whether the map key's value type is int32
func CheckMapInt32(m map[string]interface{}, key string) (int32, error) {
	if _, ok := m[key]; !ok {
		return 0, fmt.Errorf(key + " is nil")
	}

	if v, ok := m[key].(float64); ok {
		return int32(v), nil
	}
	return 0, fmt.Errorf(key + " is not int32 type")
}

// CheckMapInt check whether the map key's value type is int
func CheckMapInt(m map[string]interface{}, key string) (int, error) {
	if _, ok := m[key]; !ok {
		return 0, fmt.Errorf(key + " is nil")
	}

	if v, ok := m[key].(float64); ok {
		return int(v), nil
	}
	return 0, fmt.Errorf(key + " is not int type")
}

// CheckMapInterface check whether the map key's value type is interface{}
func CheckMapInterface(m map[string]interface{}, key string) (map[string]interface{}, error) {
	if _, ok := m[key]; !ok {
		return nil, fmt.Errorf(key + " is nil")
	}
	if v, ok := m[key].(interface{}); ok {
		return v.(map[string]interface{}), nil
	}
	return nil, fmt.Errorf(key + " is not interface{} type")
}

// CheckMapInterfaceSlice check whether the map key's value type is []interface{}
func CheckMapInterfaceSlice(m map[string]interface{}, key string) ([]interface{}, error) {
	if _, ok := m[key]; !ok {
		return nil, fmt.Errorf(key + " is nil")
	}
	if v, ok := m[key].([]interface{}); ok {
		return v, nil
	}
	return nil, fmt.Errorf(key + " is not []interface{} type")
}

// CheckMapStringSlice check whether the map key's value type is []string
func CheckMapStringSlice(m map[string]interface{}, key string) ([]string, error) {
	if _, ok := m[key]; !ok {
		return nil, fmt.Errorf(key + " is nil")
	}
	if v, ok := m[key].([]interface{}); ok {
		var stringSlice []string
		for _, s := range v {
			stringSlice = append(stringSlice, s.(string))
		}
		return stringSlice, nil
	}
	return nil, fmt.Errorf(key + " is not []interface{} type")
}

// CheckInterfaceMapInterface check whether the interface is map[string]interface{}
func CheckInterfaceMapInterface(i interface{}) (map[string]interface{}, error) {
	if v, ok := i.(interface{}); ok {
		return v.(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("%v is not interface{} type", i)
}
