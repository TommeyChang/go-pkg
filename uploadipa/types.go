package uploadipa

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// keys of the databus
const (
	dbsUsername           = "username"
	dbsPassword           = "password"
	dbsFilename           = "filename"
	dbsPackageName        = "packagename"
	dbsBytesSent          = "bytessent"
	dbsSpeed              = "speed"
	dbsBundleid           = "bundleid"
	dbsBundleVersion      = "bundleversion"
	dbsBundleShortVersion = "bundleshortversion"
	dbsSessionID          = "sessionID"
	dbsSharedSecret       = "sharedSecret"
	dbsAppleID            = "appleID"
	dbsAppName            = "appName"
	dbsIconURL            = "iconURL"
	dbsFilePath           = "filepath"
	dbsFileSize           = "filesize"
	dbsFileCheckSum       = "fileChecksum"
	dbsModifedTime        = "fileModifiedTime"
	dbsMetaSize           = "metadataSize"
	dbsMetaChecksum       = "metadataChecksum"
	dbsMetaCompressed     = "metadataCompressed"
	dbsMetadata           = "metadata"
)

type dataBus struct {
	data     map[string]interface{}
	fp       *os.File
	metadata []byte
}


func newDataBus(fp *os.File) *dataBus {
	bus := new(dataBus)
	bus.data = make(map[string]interface{}, 128)
	bus.fp = fp
	return bus
}

func (d *dataBus) set(key string, value interface{}) {
	d.data[key] = value
}

func (d *dataBus) getInt(key string) int {
	value, ok := d.data[key].(int)
	if ok {
		return value
	} else {
		return 0
	}
}

func (d *dataBus) getString(key string) string {
	value, ok := d.data[key]
	if ok {
		return fmt.Sprint(value)
	} else {
		return ""
	}

}

func (d *dataBus) getBytes(key string) []byte {
	return []byte(fmt.Sprint(d.data[key]))

}

func (d *dataBus) setFileHandler(fp *os.File) {
	d.fp = fp

}

func (d *dataBus) getFileHandler() *os.File {
	return d.fp
}

func (d *dataBus) readFileData(start, length int) []byte {
	data := make([]byte, length)
	n, _ := d.fp.ReadAt(data, int64(length))
	if n == length {
		return data
	} else {
		return nil
	}
}

func (d *dataBus) dumpMetaBuffer(data []byte) {
	d.metadata = data
}

func (d *dataBus) readMetadata(start, length int) []byte {
	return d.metadata[start:(start + length)]
}

type task struct {
	partNum  int
	fileName string
	url      string
	method   string
	offset   int
	length   int
	headers  map[string]interface{}
}


func reservation2Task(reversion interface{}) []*task {
	tasks := make([]*task, 0, 32)

	rev := reversion.(map[string]interface{})
	fileName := rev["file"].(string)
	operations := rev["operations"].([]interface{})
	for _, opInstance := range operations {
		op := opInstance.(map[string]interface{})
		r := task{fileName: fileName}
		r.offset = int(op["offset"].(float64))
		r.length = int(op["length"].(float64))
		r.headers = op["headers"].(map[string]interface{})
		r.method = op["method"].(string)
		r.url = op["uri"].(string)
		partNum := op["partNumber"].(float64)
		r.partNum = int(partNum)
		tasks = append(tasks, &r)
	}
	return tasks

}

type process struct {
	progress struct {
		sync.Mutex
		uploadSize int
		ratio      float32
		speed      int
	}
	blockNum int
	bus      *dataBus
	fileSize   int
	startTime  time.Time
	retryTimes int
}

func NewUploader(userName, appPWD, filePath string) (*process, error) {

	if bundleInfo, err := checkFile(filePath); err != nil {
		return nil, err
	} else {
		if fp, err := os.Open(filePath); err != nil {
			return nil, err
		} else {

			var bus = newDataBus(fp)
			bus.setFileHandler(fp)
			bus.set(dbsUsername, userName)
			bus.set(dbsPassword, appPWD)
			bus.set(dbsFilePath, filePath)
			bus.set(dbsPackageName, "app.itmsp")
			bus.set(dbsBytesSent, 0)
			bus.set(dbsSpeed, "N/A")
			bus.set(dbsBundleid, bundleInfo["id"])
			bus.set(dbsBundleVersion, bundleInfo["version"])
			bus.set(dbsBundleShortVersion, bundleInfo["shortVersion"])

			p := new(process)
			p.bus = bus
			p.blockNum = 4
			p.retryTimes = 3
			return p, nil
		}
	}
}

func (p *process) SetBlockNum(num int) {
	if num > 0 && num < 16 {
		p.blockNum = num
	}
}

func (p *process) SetRetryTimes(t int) {
	if t > 0 && t < 8 {
		p.retryTimes = t
	}

}

func (p *process) updateProgress(increace int) {
	p.progress.Lock()
	p.progress.uploadSize += increace
	p.progress.Unlock()

	p.progress.speed = p.progress.uploadSize / int(time.Now().Sub(p.startTime).Seconds()+1)
	p.progress.ratio = float32(p.progress.uploadSize) / float32(p.fileSize) / 100
}

func (p *process) GetProgress() float32 {
	return p.progress.ratio
}

func (p *process) GetSpeed() int {
	return p.progress.speed >> 10

}


func (p *process) GetUploadTickets() ([]interface{}, error) {
	if r, ticket, err := authForUpload(p.bus); err != nil {
		return nil, errors.New(cstServerError)
	} else {
		if r {
			return createReservation(p.bus)
		} else {
			return nil, errors.New(ticket)
		}
	}
}

func (p *process) UploadWithTickets(tickets []interface{}) (bool, string, error) {

	defer p.bus.getFileHandler().Close()

	p.fileSize = p.bus.getInt(dbsMetaSize) + p.bus.getInt(dbsFileSize)
	p.startTime = time.Now()
	if debuggerMode {
		fmt.Println("Start to upload")
	}
	for revID, rv := range tickets {
		if debuggerMode {
			fmt.Printf("send reservation %d\n", revID)
		}
		if err := p.sendData(rv); err != nil {
			return false, cstServerError, err
		}

	}
	r, info, err := uploadDone(p.bus, time.Now().Sub(p.startTime))
	if err != nil {
		return false, cstServerError, err
	}
	return r, info, err
}

func (p *process) sendData(res interface{}) error {

	var finished = struct {
		num int
		sync.Mutex
	}{}

	clockIn := func() {
		finished.Lock()
		finished.num += 1
		finished.Unlock()
	}

	tasks := reservation2Task(res)

	ctx, cancel := context.WithCancel(context.Background())

	jobChan := make(chan *task, len(tasks))

	wg := sync.WaitGroup{}
	for i := 0; i < p.blockNum; i++ {
		wg.Add(1)
		go p.sendPart(jobChan, &wg, clockIn, ctx, cancel)
	}

	for _, t := range tasks {
		jobChan <- t
	}

	close(jobChan)
	// wait for all upload ipa come back
	wg.Wait()
	cancel()
	// finishedNum must equal to tasks num
	if finished.num == len(tasks) {
		rv := res.(map[string]interface{})
		if r, info, err := commitReservation(p.bus, &rv); err != nil {
			return errors.New(cstServerError)
		} else {
			if r {
				return nil
			} else {
				return errors.New(info)
			}

		}

	}
	return errors.New(cstServerError)
}

func (p *process) sendPart(chI <-chan *task, wg *sync.WaitGroup, clockIn func(), ctx context.Context,
	cancel context.CancelFunc) {
	defer wg.Done()
	var data []byte
	// fetch data from channel
	for t := range chI {

		select {
		// if an error is occured
		case <-ctx.Done():
			return
		default:
			// check filename
			switch t.fileName {
			case "metadata.xml":
				data = p.bus.readMetadata(t.offset, t.length)
			case p.bus.getString(dbsFilename):
				data = p.bus.readFileData(t.offset, t.length)
			default:
				cancel()
				if debuggerMode {
					fmt.Printf("Filename error: name: %s, part number: %d\n", t.fileName, t.partNum)
				}
				return
			}

			failed := true
			// upload with retry
			for i := 0; i < p.retryTimes; i++ {

				rq, _ := http.NewRequest(t.method, t.url, bytes.NewBuffer(data))
				rq.Header.Add(headKeyUserAgent, headValueUserAgent)
				headers := t.headers
				for k, v := range headers {
					rq.Header.Add(k, v.(string))
				}

				if rp, err := http.DefaultClient.Do(rq); err == nil {
					if strings.Contains(rp.Status, "200") {
						p.updateProgress(t.length)
						clockIn()
						failed = false
						break
					} else {
						if debuggerMode {
							fmt.Printf("File: %s, part number: %d, tried times: %d, error: %v\n",
								t.fileName, t.partNum, i+1, err)
						}
					}
				} else {
					if debuggerMode {
						fmt.Printf("File: %s, part number: %d, tried times: %d, error: %v\n",
							t.fileName, t.partNum, i+1, err)
					}
				}
				time.Sleep(20 * time.Second)

			}
			if failed {
				cancel()
				return
			}

		}
	}

}
