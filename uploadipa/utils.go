package uploadipa

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"howett.net/plist"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func genIDString() string {
	timeList := strings.Split(time.Now().Format("2006-01-02 15:04:05.000"), " ")[:2]
	date := strings.ReplaceAll(timeList[0], "-", "")
	timeStr := strings.ReplaceAll(timeList[1], ":", "")
	timeStr = strings.Replace(timeStr, ".", "-", 1)
	return date + timeStr
}

func makeSessionDigest(sessionID, requestID, sharedSecrect string, requestCheckSum []byte) string {
	md5Worker := md5.New()
	md5Worker.Write([]byte(sessionID))
	md5Worker.Write(requestCheckSum)
	md5Worker.Write([]byte(requestID))
	md5Worker.Write([]byte(sharedSecrect))
	return hex.EncodeToString(md5Worker.Sum(nil))

}

func genMetadata(bus *dataBus) {

	state, _ := bus.getFileHandler().Stat()
	bus.set(dbsFilename, state.Name())
	bus.set(dbsFileSize, int(state.Size()))
	bus.set(dbsModifedTime, int(state.ModTime().UnixNano() / 1e6))
	fileCheckSum := getFileMD5(bus.getFileHandler())
	bus.set(dbsFileCheckSum, fileCheckSum)

	data := []byte(cstMetadata)
	data = bytes.ReplaceAll(data, []byte("APPLE_ID"), bus.getBytes(dbsAppleID))
	data = bytes.ReplaceAll(data, []byte("BUNDLE_SHORT_VERSION"), bus.getBytes(dbsBundleShortVersion))
	data = bytes.ReplaceAll(data, []byte("BUNDLE_VERSION"), bus.getBytes(dbsBundleVersion))
	data = bytes.ReplaceAll(data, []byte("BUNDLE_IDENTIFIER"), bus.getBytes(dbsBundleid))
	data = bytes.ReplaceAll(data, []byte("FILE_SIZE"), bus.getBytes(dbsFileSize))
	data = bytes.ReplaceAll(data, []byte("FILE_NAME"), bus.getBytes(dbsFilename))
	data = bytes.ReplaceAll(data, []byte("MD5"), bus.getBytes(dbsFileCheckSum))

	lenData := len(data)
	bus.set(dbsMetaSize, lenData)
	bus.set(dbsMetaChecksum, getBytesMD5(data))
	bus.set(dbsMetadata, string(data))

	bus.dumpMetaBuffer(data)
	metaCompressed := gzBase64(data)
	bus.set(dbsMetaCompressed, metaCompressed)
	return
}

func getFileMD5(fp *os.File) string {
	md5h := md5.New()
	io.Copy(md5h, fp)
	return hex.EncodeToString(md5h.Sum(nil))
}

func getBytesMD5(data []byte) string {
	md5h := md5.New()
	md5h.Write(data)
	return hex.EncodeToString(md5h.Sum(nil))
}

func gzBase64(data []byte) string {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	defer gz.Close()
	gz.Write(data)
	gz.Flush()
	return base64.StdEncoding.EncodeToString(buffer.Bytes())

}

func sendData(bus *dataBus, tasks []*task) error {

	var data []byte
	for _, t := range tasks {
		if t.fileName == "metadata.xml" {
			data = bus.readMetadata(t.offset, t.length)
		} else if t.fileName == bus.getString(dbsFilename) {
			data = bus.readFileData(t.offset, t.length)
		} else {
			return errors.New("unknown file")
		}

		if rq, err := http.NewRequest(t.method, t.url, bytes.NewBuffer(data)); err != nil {
			return err
		} else {
			rq.Header.Add(headKeyUserAgent, headValueUserAgent)
			for k, v := range t.headers {
				rq.Header.Add(k, v.(string))
			}

			if rp, err := http.DefaultClient.Do(rq); err != nil {
				return err
			} else {
				if !strings.Contains(rp.Status, "200") {
					return errors.New(rp.Status)
				}
			}

		}

	}
	return nil

}


func ValidateAccount(account, appPWD string) (bool, error) {
	reqID := genIDString()

	rqData := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  cstAuthForSessioon,
		"id":      reqID,
		"params": map[string]string{
			"Username": account,
			"Password": appPWD,
		},
	}

	rqJSON, _ := json.Marshal(rqData)

	rq, _ := http.NewRequest("POST", cstProduceURL, bytes.NewBuffer(rqJSON))
	rq.Header.Set(headKeyContentType, headValueContenType)
	rq.Header.Set(headKeyUserAgent, headValueUserAgent)

	if rp, err := http.DefaultClient.Do(rq); err != nil {
		return false, err
	} else {
		if body, err := ioutil.ReadAll(rp.Body); err != nil {
			return false, err
		} else {
			defer rp.Body.Close()
			var resultRaw map[string]interface{}
			if err := json.Unmarshal(body, &resultRaw); err != nil {
				return false, err
			}
			sessionData := resultRaw["result"].(map[string]interface{})

			_, okID := sessionData["SessionId"]
			_, okS := sessionData["SharedSecret"]
			if okID && okS {
				return true, nil
			} else {
				return false, nil

			}
		}
	}
}

func checkFile(fileName string) (map[string]string, error) {
	zr, err := zip.OpenReader(fileName)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return getBundleAttr(zr)

}


func getBundleAttr(zr *zip.ReadCloser) (map[string]string, error) {
	for _, file := range zr.File {
		if  strings.Contains(file.Name, "Info.plist") {
			nameSplit := strings.Split(file.Name, string(os.PathSeparator))
			if strings.Contains(nameSplit[len(nameSplit)-2], ".app") {
				if fp, err := file.Open(); err == nil {
					if bs, err := ioutil.ReadAll(fp); err == nil {
						var result map[string]interface{}
						if _, err := plist.Unmarshal(bs, &result); err == nil {
							fp.Close()
							bundleID, okB := result["CFBundleIdentifier"]
							bundleVersion, okV := result["CFBundleVersion"]
							bundleShortVersion, okSV := result["CFBundleShortVersionString"]
							bundleName, okN := result["CFBundleDisplayName"]
							if okB && okSV && okV && okN {
								result :=  map[string]string{
									"id": bundleID.(string),
									"version": bundleVersion.(string),
									"shortVersion": bundleShortVersion.(string),
									"name": bundleName.(string),
								}
								return result, nil
								break
							}
						} else {
							fp.Close()
						}

					}

				}

			}

		}
	}
	return nil, errors.New("no Info.plist file found")
}

func authForUpload(bus *dataBus) (bool, string, error) {
	if debuggerMode {
		fmt.Println("Auth for session")
	}
	r, info, err := authForSession(bus)
	if err != nil {
		return false, "", err
	}
	if !r {
		return false, info, nil
	}
	if debuggerMode {
		fmt.Println("lookup software for bundle id")
	}
	r, info, err = lookupSoftwareForBundleID(bus)
	if err != nil {
		return false, "", err
	}
	if !r {
		return false, info, nil
	}
	genMetadata(bus)
	if err != nil {
		return false, "", err
	}
	if debuggerMode {
		fmt.Println("validate metadata")
	}
	r, info, err = validMetadata(bus)
	if err != nil {
		return false, "", err
	}
	if !r {
		return false, info, err
	}
	if debuggerMode {
		fmt.Println("validate assets")
	}
	r, info, err = validAssets(bus)
	if err != nil {
		return false, "", err
	}
	if !r {
		return false, info, err
	}
	if debuggerMode {
		fmt.Println("client checksum")
	}
	r, info, err = clientCheckSum(bus)
	if err != nil {
		return false, "", err
	}
	if !r {
		return false, info, err
	}
	return true, "", nil

}
