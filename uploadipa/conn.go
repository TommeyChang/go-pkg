package uploadipa

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func makeProducerServiceReq(bus *dataBus, method string, param map[string]interface{}) (*map[string]interface{}, error) {

	reqID := genIDString()

	rqData := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      reqID,
		"params":  param,
	}

	if rqJSON, err := json.Marshal(rqData); err != nil {
		return nil, err
	} else {

		if rq, err := http.NewRequest("POST", cstProduceURL, bytes.NewBuffer(rqJSON)); err != nil {
			return nil, err
		} else {
			rq.Header.Set(headKeyContentType, headValueContenType)
			rq.Header.Set(headKeyUserAgent, headValueUserAgent)

			sessionID := bus.getString(dbsSessionID)

			if sessionID != "" {
				md5h := md5.New()
				md5h.Write(rqJSON)
				sessionDigest := makeSessionDigest(
					sessionID, reqID, bus.getString(dbsSharedSecret), md5h.Sum(nil))
				rq.Header.Add(xRequestID, reqID)
				rq.Header.Add(xSessionDigest, sessionDigest)
				rq.Header.Add(xSessionID, bus.getString(dbsSessionID))
				rq.Header.Add(xSessionVersion, "2")
			}

			if rp, err := http.DefaultClient.Do(rq); err != nil {
				return nil, err
			} else {
				if body, err := ioutil.ReadAll(rp.Body); err != nil {
					return nil, err
				} else {
					defer rp.Body.Close()
					var resultRaw map[string]interface{}
					if err := json.Unmarshal(body, &resultRaw); err != nil {
						return nil, err
					}
					sessionDataRaw := resultRaw["result"].(map[string]interface{})
					return &sessionDataRaw, nil
				}

			}

		}
	}

}


func makeSoftwareServiceReq(bus *dataBus, method string, param map[string]interface{}) (*map[string]interface{}, error) {

	reqID := genIDString()

	rqData := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      reqID,
		"params":  param,
	}

	if rqJSON, err := json.Marshal(rqData); err != nil {
		return nil, err
	} else {

		if rq, err := http.NewRequest("POST", cstSoftURL, bytes.NewBuffer(rqJSON)); err != nil {
			return nil, err
		} else {
			rq.Header.Set(headKeyContentType, headValueContenType)
			rq.Header.Set(headKeyUserAgent, headValueUserAgent)

			sessionID := bus.getString(dbsSessionID)

			if sessionID != "" {
				md5h := md5.New()
				md5h.Write(rqJSON)
				sessionDigest := makeSessionDigest(
					sessionID, reqID, bus.getString(dbsSharedSecret), md5h.Sum(nil))
				rq.Header.Add(xRequestID, reqID)
				rq.Header.Add(xSessionDigest, sessionDigest)
				rq.Header.Add(xSessionID, bus.getString(dbsSessionID))
				rq.Header.Add(xSessionVersion, "2")

			}
			if rp, err := http.DefaultClient.Do(rq); err != nil {
				return nil, err
			} else {
				if body, err := ioutil.ReadAll(rp.Body); err != nil {
					return nil, err
				} else {
					defer rp.Body.Close()
					var resultRaw map[string]interface{}
					if err := json.Unmarshal(body, &resultRaw); err != nil {
						return nil, err
					}
					sessionDataRaw := resultRaw["result"].(map[string]interface{})
					return &sessionDataRaw, nil
				}

			}
		}

	}
}
