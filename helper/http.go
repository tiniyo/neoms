package helper

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/neoms/logger"
)

/*
var hbConnections []*resty.Client
var rrConnections []*resty.Client
var sipConnections []*resty.Client
var numberConnections []*resty.Client
var genConnections []*resty.Client

var hbindex = 0
var rrindex = 0
var sipindex = 0
var numberindex = 0
var genindex = 0

var maxsize = 100
var mutext sync.Mutex

func InitHttpConnPool() {
	CreateHttpClientPool("heartbeat")
	CreateHttpClientPool("rateroute")
	CreateHttpClientPool("sip")
	CreateHttpClientPool("number")
	CreateHttpClientPool("gen")
}

	mutext.Unlock()
}
*/
func createTransport(localAddr net.Addr) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   1000,
		MaxConnsPerHost:       1000,
	}
}

var reusableConnection *resty.Client

func getConnection2() *resty.Client {
	if reusableConnection != nil {
		return reusableConnection
	}
	reusableConnection := resty.New()
	transport := createTransport(nil)
	reusableConnection.SetTransport(transport)
	return reusableConnection
}

/*
func getConnection(clientType string) *resty.Client {
	var conn *resty.Client
	mutext.Lock()
	switch clientType {
	case "heartbeat":
		hbindex = hbindex % maxsize
		conn = hbConnections[hbindex]
		hbindex++
	case "rateroute":
		rrindex = rrindex % maxsize
		conn = rrConnections[rrindex]
		rrindex++
	case "sip":
		sipindex = rrindex % maxsize
		conn = sipConnections[sipindex]
		sipindex++
	case "number":
		numberindex = rrindex % maxsize
		conn = numberConnections[numberindex]
		numberindex++
	default:
		genindex = rrindex % maxsize
		conn = genConnections[genindex]
		genindex++
	}
	mutext.Unlock()
	return conn
}
*/

func Get(callSid string, restData map[string]interface{}, urls string) (int, []byte, error) {
	var clientType = getClientType(urls)
	var restClient = getConnection2()
	if clientType == "heartbeat" || clientType == "rateroute" || clientType == "sip" || clientType == "number" {
		restClient.SetBasicAuth("FAIeewew232324", "KO66!RWHFh>9J!~;oFCV[lPN0")
	} else {
		u, err := url.Parse(urls)
		if err != nil {
			logger.UuidLog("Err", callSid, err.Error())
			return 400, nil, err
		}
		if u.User != nil {
			if pwd, ok := u.User.Password(); ok {
				restClient.SetBasicAuth(u.User.Username(), pwd)
				baseUrlStr := fmt.Sprint(u.Scheme, "://", u.Host, u.Path)
				base, err := url.Parse(baseUrlStr)
				if err != nil {
					return 400, nil, err
				}
				urls = fmt.Sprint(base.ResolveReference(u))
				logger.UuidLog("Info", callSid, urls)
			}
		}
	}
	if restData == nil {
		resp, err := restClient.R().
			EnableTrace().
			SetHeader("Accept", "application/json").
			Get(urls)
		logger.UuidHttpLog(callSid, resp)
		if resp == nil {
			return 400, nil, err
		}
		return resp.StatusCode(), resp.Body(), err
	}
	reqData := make(map[string]string)
	for key, value := range restData {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)
		reqData[strKey] = strValue
	}
	resp, err := restClient.R().
		SetQueryParams(reqData).
		EnableTrace().
		SetHeader("Accept", "application/json").
		Get(urls)
	logger.UuidHttpLog(callSid, resp)

	// Close the connection to reuse it
	if resp == nil {
		return 400, nil, err
	}
	return resp.StatusCode(), resp.Body(), err
}

func Post(callSid string, restData map[string]interface{}, urls string) (int, []byte, error) {
	var clientType = getClientType(urls)
	var restClient = getConnection2()
	if clientType == "heartbeat" || clientType == "rateroute" || clientType == "sip" || clientType == "number" {
		restClient.SetBasicAuth("FGHGHGHGHGVDFDFD", "KO66!RWHFh>9J!~;oFCV[lPN0")
	} else {
		u, err := url.Parse(urls)
		if err != nil {
			logger.UuidLog("Err", callSid, err.Error())
			return 400, nil, err
		}
		if u.User != nil {
			logger.UuidLog("Info", callSid, "found username and password in url ")
			if pwd, ok := u.User.Password(); ok {
				restClient.SetBasicAuth(u.User.Username(), pwd)
				baseUrlStr := fmt.Sprint(u.Scheme, "://", u.Host, u.Path)
				base, err := url.Parse(baseUrlStr)
				if err != nil {
					return 400, nil, err
				}
				urls = fmt.Sprint(base.ResolveReference(u))
				logger.UuidLog("Info", callSid, urls)
			}
		}
	}
	if restData == nil {
		resp, err := restClient.R().
			EnableTrace().
			SetHeader("Accept", "application/json").
			Post(urls)
		logger.UuidHttpLog(callSid, resp)
		if resp == nil {
			return 400, nil, err
		}
		return resp.StatusCode(), resp.Body(), err
	}
	resp, err := restClient.R().
		EnableTrace().
		SetBody(restData).
		SetHeader("Accept", "application/json").
		Post(urls)
	logger.UuidHttpLog(callSid, resp)
	if resp == nil {
		return 400, nil, err
	}
	return resp.StatusCode(), resp.Body(), err
}

func getClientType(url string) string {
	var clientType = ""
	if strings.Contains(url, "HeartBeat") {
		clientType = "heartbeat"
	} else if strings.Contains(url, "RateRoutes") {
		clientType = "rateroute"
	} else if strings.Contains(url, "PhoneNumbers") {
		clientType = "number"
	} else if strings.Contains(url, "Sips") {
		clientType = "sip"
	} else {
		clientType = "generic"
	}
	return clientType
}
