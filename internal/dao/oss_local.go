package dao

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/sirupsen/logrus"
)

func newLocalossServent() (*localossServant, versionInfo) {
	savePath, err := filepath.Abs(conf.LocalOSSSetting.SavePath)
	if err != nil {
		logrus.Fatalf("get localOSS save path err: %v", err)
	}

	obj := &localossServant{
		savePath: savePath + "/" + conf.LocalOSSSetting.Bucket + "/",
		domain:   getOssDomain(),
	}
	return obj, obj
}

func (s *localossServant) name() string {
	return "LocalOSS"
}

func (s *localossServant) version() *semver.Version {
	return semver.MustParse("v0.1.0")
}

func (s *localossServant) PutObject(objectKey string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	saveDir := s.savePath + filepath.Dir(objectKey)
	err := os.MkdirAll(saveDir, 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	savePath := s.savePath + objectKey
	writer, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	written, err := io.Copy(writer, reader)
	if err != nil {
		return "", err
	}
	if written != objectSize {
		os.Remove(savePath)
		return "", errors.New("put object not complete")
	}

	return s.domain + objectKey, nil
}

func (s *localossServant) SignURL(objectKey string, expiredInSec int64) (string, error) {
	if expiredInSec < 0 {
		return "", fmt.Errorf("invalid expires: %d, expires must bigger than 0", expiredInSec)
	}
	expiration := time.Now().Unix() + expiredInSec

	// Fixed: Just make things simple and simple now so return an veiry simple sign url.
	// Maybe make another process logic for sign url in future but not now.
	uri := fmt.Sprintf("%s%s?expired=%d", s.domain, objectKey, expiration)
	return uri, nil
}

func (s *localossServant) ObjectURL(objetKey string) string {
	return s.domain + objetKey
}

func (s *localossServant) ObjectKey(objectUrl string) string {
	return strings.Replace(objectUrl, s.domain, "", -1)
}
