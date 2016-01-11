package reaper

import (
	"calculator/config"
	"calculator/sub"
	"calculator/kafka"
	"calculator/tsdb"
	"encoding/json"
	"fmt"
	"gopkg.in/tomb.v1"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"utils"
	"utils/logger"
	"utils/sorted"
)

/*
	deserialize data from DomainBindUrl.
	get data from http://tsdb2.hy01.internal.wandoujia.com.

*/
type DomainBindInfo struct {
	Data NodeBindInfo `json:"data"`
}
type BindInfo struct {
	Path   string `json:"path"`
	Domain string `json:"domain"`
}
type NodeBindInfo map[string][]BindInfo

/*
	Data struct for specified domain and tags.
*/
type specifiledDataInfo struct {
	mutex sync.Mutex
	sTime int64
	cTime int64
	data  map[string]*TsdbPutInfo
}
type TsdbPutInfo struct {
	TimeStamp  int64             `json:"timestamp"`
	Metric     string            `json:"metric"`
	Value      float64           `json:"value"`
	Tags       map[string]string `json:"tags"`
	Aggregated bool              `json:"-"`
}

/*
	Load yaml config file and save as blow.
*/
type AnalyzeTemplate struct {
	Domain string   `yaml:"domain"`
	Tags   []string `yaml:"tags"`
}

type domainPathList struct {
	rmux sync.RWMutex
	data map[string]int
}

type domainPathData struct {
	domain   string
	dataInfo *specifiledDataInfo
	pathList *domainPathData
}

type AggregationData struct {
	domain string
	data   map[string]string
}
type domainMetricTsdbMap map[string]*TsdbPutInfo

type Raper struct {
	mutexDomains     sync.RWMutex
	mutexBindPath    sync.RWMutex
	mutexTags        sync.RWMutex
	mutexDomainMap   sync.RWMutex
	domainBindPath   map[string][]string
	aggregatedPath   map[string][]string
	totalDomainTags  map[string][]string
	domainsChanMap   map[string]chan domainMetricTsdbMap
	domains          []string
	subMessageChan   <-chan string
	aggreMessageChan chan *AggregationData
	producer 		 *kafka.Producer
	tsdb             *tsdb.TsdbPipeLine
	subscriber       *sub.Subscriber
	tomb.Tomb
}

// start service.
func StartService(file string) {

	r := &Raper{
		// cache domain binding data.
		domainBindPath: make(map[string][]string),
		// save domain Tags for each domain
		totalDomainTags: make(map[string][]string),
		// save each domain's message.
		aggreMessageChan: make(chan *AggregationData, 4096),
		// chan used to upload data message to tsdb
		aggregatedPath: make(map[string][]string),
		// tsdb pipe line
		tsdb: tsdb.NewPipeLine(),
	}

	defer r.Close()
	r.subscriber = sub.NewSub()
	r.subscriber.GetMessage()

	r.producer = kafka.NewPub()
	go r.producer.PublishToKafka()

	domainBindChan := getDomainPath()

	c := <-domainBindChan
	r.saveBindData(c)
	logger.Println("finish saving bind path for the first time")
	logger.Println("finish connecting to tsdb for the first time")
	go r.distDomainMertic()
	go r.createMetricTags(file)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case c := <-domainBindChan:
			go r.saveBindData(c)
		case m := <-r.subscriber.LogsChan:
			r.analyzeMessage(m)
		case _ = <-signals:
			r.Kill(nil)
			r.Close()
		}
	}

}

func (r *Raper) Close() {
	logger.Println("close all sockets here and exit")
	r.subscriber.Close()
	r.tsdb.Close()
	os.Exit(0)
}

func getDomainPath() <-chan DomainBindInfo {
	var domainBindChan chan DomainBindInfo = make(chan DomainBindInfo)
	var resp *http.Response
	var err error
	go func() {
		for {
			resp, err = http.Get(config.DomainBindUrl)
			if err != nil {
				logger.Printf("get domain bindings failed with url: %s", config.DomainBindUrl)
				time.Sleep(time.Second)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)

			if err == nil {
				dp := DomainBindInfo{}
				err = json.Unmarshal(body, &dp)
				if err != nil {
					logger.Printf("unmarshal domain binding json data failed: %+v\n", err.Error())
				} else {
					domainBindChan <- dp
					time.Sleep(time.Second * 15)
				}

			} else {
				time.Sleep(time.Second)
			}
		}
	}()

	return domainBindChan
}

func (r *Raper) createMetricTags(file string) {
	readLoad := make(chan struct{}, 1)
	readLoad <- struct{}{}
	go func() {
		for {
			fi, err := os.Stat(file)
			if err == nil {
				modTime := fi.ModTime()
				if time.Since(modTime) < time.Duration(30)*time.Second {
					readLoad <- struct{}{}
				}
			}
			time.Sleep(time.Duration(25) * time.Second)
		}
	}()
	for {
		select {
		case <-readLoad:
			var out []AnalyzeTemplate
			var defaultLen int = 512
			var content []byte
			readonce := make([]byte, defaultLen)
			f, err := os.Open(file)
			if err == nil {
				for {
					length, err := f.Read(readonce)

					content = append(content, readonce[0:length]...)

					if err == io.EOF {
						break
					}

				}
				err = yaml.Unmarshal(content, &out)
				if err != nil {
					logger.Errorf("%s is illegal yaml format! %+v", content, err)
				}

			}
			if err != nil {
				time.Sleep(3 * time.Second)
				break
			}

			logger.Printf("load yaml file ok, result: %#v\n", out)
			r.mutexTags.Lock()
			for _, value := range out {
				r.totalDomainTags[value.Domain] = value.Tags
			}
			r.mutexTags.Unlock()

		case <-r.Dying():
			return
		}
	}

}

func (r *Raper) saveBindData(c DomainBindInfo) {

	var newBindData map[string][]string = make(map[string][]string)

	for _, v1 := range c.Data {
		for _, v2 := range v1 {
			domain := v2.Domain
			path := v2.Path
			if ok := utils.StrInStrings(path, newBindData[domain]); !ok {
				newBindData[domain] = append(newBindData[domain], path)
			}

		}
	}

	r.mutexBindPath.Lock()
	if ok := utils.MapStrStringsCmp(r.domainBindPath, newBindData); !ok {
		logger.Println("save new domain bind data")

		r.domainBindPath = newBindData

	}
	r.mutexBindPath.Unlock()

	var domains []string

	for k := range r.domainBindPath {
		domains = append(domains, k)
	}
	for _, d := range domains {
		r.mutexBindPath.RLock()
		a := make([]string, len(r.domainBindPath[d]))
		copy(a, r.domainBindPath[d])
		r.mutexBindPath.RUnlock()
		bcc := sorted.ByCharCount(a, "/")
		sort.Sort(bcc)
		r.mutexBindPath.Lock()
		r.domainBindPath[d] = bcc.List
		r.mutexBindPath.Unlock()
	}

}

func (r *Raper) analyzeMessage(JsonObj sub.LogJsonObj) {
	dataMap := make(map[string]string)

	var method, protocol, path, code, length string

	line := strings.SplitN(JsonObj.LineText, "\"", 3)

	if len(line) < 3 {
		logger.Printf("illegal line:%+v from:%+v", JsonObj.LineText, JsonObj.HostName)
		return
	}
	right := strings.Replace(line[2], "\"", "", -1)
	logs := strings.Split(right, " ")
	logs = utils.TripStringsBlank(logs)
	middle := strings.Split(line[1], " ")
	if len(logs) < 2 {
		logger.Printf("illegal right:%+v from:%+v", JsonObj.LineText, JsonObj.HostName)
		return
	} else if len(middle) < 3 {
		logger.Printf("illegal middle:%+v from:%+v", JsonObj.LineText, JsonObj.HostName)
		return
	}

	method, protocol = middle[0], middle[len(middle)-1]
	if method == "-" {
		return
	}

	path = strings.Split(middle[1], "?")[0]
	code, length = logs[0], logs[1]

	//get the http code that bigger than 499 
	v_code, _ := utils.ParseInt64(code)
	if v_code >= 499 {
		v_data, _ := json.Marshal(JsonObj)
		r.producer.MessageChan <- v_data
	}

	for _, leftPart := range logs[2:] {
		if isKeyValue := strings.Count(leftPart, "="); isKeyValue == 0 {
			continue
		}

		kvPair := strings.Split(leftPart, "=")
		if kvPair[1] == "-" {
			continue
		}
		dataMap[kvPair[0]] = kvPair[1]
	}
	if _, ok := dataMap["host"]; !ok {
		// logger.Printf("illegal domain:%+v from:%+v:%+v", v.LineText, v.HostName, v.FileName)
		return
	}

	dataMap["method"] = method
	dataMap["proto"] = protocol
	dataMap["path"] = path
	dataMap["code"] = code
	dataMap["length"] = length
	dataMap["domain"] = dataMap["host"]
	dataMap["source"] = JsonObj.HostName
	delete(dataMap, "host")

	dataMap = utils.CheckLegalChar(dataMap)
	r.aggreMessageChan <- &AggregationData{domain: dataMap["domain"], data: dataMap}
}

func (r *Raper) getDomainTags(domain string) (configTags []string) {

	r.mutexTags.RLock()
	if _, ok := r.totalDomainTags[domain]; !ok {
		configTags = config.UrlDefaultTags
	} else {
		configTags = r.totalDomainTags[domain]
	}
	r.mutexTags.RUnlock()

	return

}

func (r *Raper) updateDomainMetric(domain string, c chan domainMetricTsdbMap) {

	stamp, err := utils.CurrentStamp(config.UploadFrequency)
	if err != nil {
		logger.Fatalf("UploadFrequency must be integer times of 60, err:%d\n", config.UploadFrequency)
	}

	dataInfo := &specifiledDataInfo{
		data:  make(map[string]*TsdbPutInfo),
		sTime: stamp,
	}

	pathList := &domainPathList{
		data: make(map[string]int),
	}

	go r.createAndUploadData(domain, pathList, dataInfo)

	for {
		select {
		case tsdb := <-c:
			for k, v := range tsdb {

				path := v.Tags["path"]

				if v.Metric == config.UrlCodeMetric {
					pathList.rmux.Lock()
					if _, ok := pathList.data[path]; !ok {
						pathList.data[path] = 1
					} else {
						pathList.data[path]++
					}
					pathList.rmux.Unlock()

				}

				dataInfo.mutex.Lock()
				if _, ok := dataInfo.data[k]; !ok {
					dataInfo.data[k] = v
					dataInfo.data[k].TimeStamp = dataInfo.sTime
				} else {
					dataInfo.data[k].Value += v.Value
				}
				dataInfo.mutex.Unlock()
			}
		case <-r.Dying():
			return
		}
	}
}

func (r *Raper) distDomainMertic() {

	r.domainsChanMap = make(map[string]chan domainMetricTsdbMap)
	for {
		select {
		case AD := <-r.aggreMessageChan:
			domain, data := AD.domain, AD.data
			dataInfo := make(domainMetricTsdbMap)
			r.mutexDomains.Lock()
			if !utils.StrInStrings(domain, r.domains) {
				r.domains = append(r.domains, domain)
			}
			r.mutexDomains.Unlock()

			key, tags := "", make(map[string]string)

			configTags := r.getDomainTags(domain)

			tags, key = utils.TagsToKey(configTags, data)

			length, err := utils.ParseFloat64(data["length"])
			if err != nil {
				logger.Printf("parse length error: %+v\n", data)
				break
			}

			requestTime, err := utils.ParseFloat64(data["reqtime"])
			if err != nil {
				logger.Printf("request time parse error: %+v", data)
				requestTime = 0
			}

			upstreamTime, err := utils.ParseFloat64(data["upstream_resptime"])
			if err != nil {
				upstreamTime = requestTime
			}

			/*
				code 408 means server wait for client request timeout, so ignore here.
			*/
			if c := data["code"]; c == "408" || c == "499" && requestTime < config.Code499Timeout {
				break
			}

			for _, v := range config.TotalUrlMetric {
				k := key + "|" + v
				if _, ok := dataInfo[k]; !ok {
					dataInfo[k] = &TsdbPutInfo{
						Metric: v,
						Tags:   tags,
					}
				}
			}
			dataInfo[key+"|"+config.UrlQpsMetric].Value += 1 / float64(config.UploadFrequency)
			dataInfo[key+"|"+config.UrlCodeMetric].Value += 1
			dataInfo[key+"|"+config.UrlUpstreamMetric].Value += upstreamTime * 1000
			dataInfo[key+"|"+config.UrlTimeMetric].Value += requestTime * 1000
			dataInfo[key+"|"+config.UrlTrafficMetric].Value += length / float64(config.UploadFrequency)

			var channel chan domainMetricTsdbMap
			r.mutexDomainMap.RLock()
			_, ok := r.domainsChanMap[domain]
			r.mutexDomainMap.RUnlock()
			if !ok {
				r.mutexDomainMap.Lock()
				_, ok := r.domainsChanMap[domain]
				if !ok {
					r.domainsChanMap[domain] = make(chan domainMetricTsdbMap)
					go r.updateDomainMetric(domain, r.domainsChanMap[domain])
				}
				channel = r.domainsChanMap[domain]
				r.mutexDomainMap.Unlock()

			} else {
				r.mutexDomainMap.RLock()
				channel = r.domainsChanMap[domain]
				r.mutexDomainMap.RUnlock()
			}

			channel <- dataInfo
		case <-r.Dying():
			return
		}
	}

}

func (r *Raper) createAndUploadData(domain string, pathList *domainPathList, dataInfo *specifiledDataInfo) {
	hostName, _ := os.Hostname()
	var conftags []string
	for {
		select {
		case <-time.After(time.Duration(dataInfo.sTime-utils.CurrentMilliSecond()+int64(config.UploadFrequency)*1000) * time.Millisecond):
			dataInfo.mutex.Lock()
			tsdbMap := dataInfo.data
			dataInfo.data = make(map[string]*TsdbPutInfo)
			dataInfo.sTime, _ = utils.CurrentStamp(config.UploadFrequency)
			dataInfo.mutex.Unlock()

			pathList.rmux.Lock()
			countMap := pathList.data
			pathList.data = make(map[string]int)
			pathList.rmux.Unlock()

			finalTsdbMap := make(domainMetricTsdbMap)
			paths := utils.UTF8Filter(AggregationPath(countMap))
			r.mutexBindPath.RLock()
			domainBinds := r.domainBindPath[domain]
			r.mutexBindPath.RUnlock()
			paths = utils.AppendListToList(domainBinds, paths)
			conftags = r.getDomainTags(domain)

			var total float64
			for key, tsdb := range tsdbMap {

				oldPath := tsdb.Tags["path"]

				tsdb.Tags["path"], _ = utils.FindConfigPath(oldPath, paths)
				_, newKey := utils.TagsToKey(conftags, tsdb.Tags)
				parts := strings.Split(key, "|")

				lastPart := parts[len(parts)-1]
				finalKey := newKey + "|" + lastPart

				if lastPart == config.UrlCodeMetric {
					total += tsdb.Value
				}

				if _, ok := finalTsdbMap[finalKey]; ok {
					finalTsdbMap[finalKey].Value += tsdbMap[key].Value
				} else {
					finalTsdbMap[finalKey] = tsdbMap[key]
				}
			}
			logger.Printf("domain: %s, total: %d\n", domain, int64(total))
			go func() {
				for _, v := range finalTsdbMap {
					var data string
					intPart, frac := math.Modf(v.Value)
					if frac == 0 {
						data = fmt.Sprintf("put %s %d %d consumer=%s", v.Metric, v.TimeStamp, int64(intPart), hostName)
					} else {
						data = fmt.Sprintf("put %s %d %.2f consumer=%s", v.Metric, v.TimeStamp, v.Value, hostName)
					}
					for key, value := range v.Tags {
						data += fmt.Sprintf(" %s=%s", key, value)
					}
					// put test.url.code 1440492540000 2000 domain=api.wandoujia.com path=/sre-test code=200
					data += "\n"
					r.tsdb.MessageChan <- data
				}
			}()
		case <-r.Dying():
			return
		}
	}
}
