package backend_process

import (
	"dujiacheng.jason/tail-based-sampling-go/src/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type void struct{}

var (
	traceChucksumMap map[string]string
	traceChunksumOriginMap map[string]string
	file *os.File
)

func init() {
	file, _ = os.OpenFile("op.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file.Close()
	log.SetPrefix("TRACE: ")
	traceChucksumMap = make(map[string]string)
}

func Start() {
	go Run()
}

func Run() {
	var traceIdBatch *TraceIdBatch
	ports := []string{"8000", "8001"}
	for {
		traceIdBatch = getFinishedBatch()
		if traceIdBatch == nil {
			if isFinished() {
				if sendCheckSum() {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}
		// 应该不至于有重的吧...会有的！因为是3个batch里搜的？比如某个traceId在batch_6，会从456，567，678中获取wrongTraceList来合并
		fmap := make(map[string]map[string]void) //Map<String, Set<String>> map = new HashMap<>(); 这是总的wrongTrace的map，key为traceId! processMap是每一次查的，要汇入到fmap中
		batchPos := traceIdBatch.batchPos
		for _, port := range ports {
			traceIdBatchStr, err := json.Marshal(traceIdBatch.traceIdList)
			if err != nil {
				log.Fatalln("traceIdBatch json marshal fail,err.", err)
			}
			processMap := getWrongTrace(string(traceIdBatchStr), port, batchPos)
			if processMap != nil {
				for traceId, wrongTraceList := range processMap {
					//fmap[traceId] = make(map[string]void) // 巨型bug！这样的话岂不是每个port都会重置fmap[traceId]，会导致结果只有一半！
					if fmap[traceId] == nil{
						fmap[traceId] = make(map[string]void)
					}
					for _, v := range wrongTraceList {
						fmap[traceId][v] = void{}
					}
				}
			}
		}
		//log.Printf("getWrong:%d,traceIdsize:%d,result:%d", batchPos, len(traceIdBatch.traceIdList), len(fmap))

		for traceId, spanSet := range fmap {
			spans := sortAndJoin(spanSet)
			traceChucksumMap[traceId] = utils.MD5(spans)
		}
	}
}

func sortAndJoin(set map[string]void) string {
	// convert to list
	list := make([]string,0,len(set))
	for k, _ := range set {
		list = append(list, k)
	}
	// Sort by startTime
	sort.Slice(list, func(i, j int) bool {
		return getStartTime(list[i]) < getStartTime(list[j])
	})
	// Join
	sb := strings.Builder{}
	for _, v := range list {
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	content := sb.String()

	// 测试是否转义有问题
	content = strings.Replace(content, "\\u003c", "<", -1)
	content = strings.Replace(content, "\\u003e", ">", -1)
	content = strings.Replace(content, "\\u0026", "&", -1)

	return content
}

func sendCheckSum() bool {
	file5, _ := os.OpenFile("checksum_Origin.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file5.Close()
	for _,v := range traceChunksumOriginMap{
		fmt.Fprintln(file5,v)
	}

	jsonStr,err := json.Marshal(traceChucksumMap)
	// test
	file3, _ := os.OpenFile("checksum.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file.Close()
	fmt.Fprintln(file3,string(jsonStr))



	if err != nil {
		log.Fatalln("sendCheckSum.err", err)
		return false
	}
	reqUrl := fmt.Sprintf("http://localhost:%s/api/finished", utils.DataSourcePort)
	fmt.Println(reqUrl)
	fmt.Println(utils.DataSourcePort)
	time.Sleep(5 * time.Second)
	_, err = http.PostForm(reqUrl, url.Values{"result": {string(jsonStr)}})
	if err != nil {
		log.Fatalln("sendCheckSum err", err)
		return false
	}
	return true
}

func getStartTime(span string) int64 {
	if span != "" {
		cols := strings.Split(span, "|")
		if len(cols) > 8 {
			t, err := strconv.Atoi(cols[1])
			if err != nil{
				fmt.Println("parse startTime err.",err)
			}
			return int64(t)
		}
	}
	return -1
}

func getWrongTrace(traceIdListStr string, port string, batchPos int) map[string][]string {
	reqUrl := fmt.Sprintf("http://localhost:%s/getWrongTrace", port)
	//fmt.Println("getWrongTrace.req:",url.Values{"traceIdList": {traceIdListStr}, "batchPos": {strconv.Itoa(batchPos)}})
	resp, err := http.PostForm(reqUrl, url.Values{"traceIdList": {traceIdListStr}, "batchPos": {strconv.Itoa(batchPos)}})
	if resp == nil || err != nil {
		log.Fatalln("getWrongTrace err", err)
	}

	body, _ := ioutil.ReadAll(resp.Body) // 短resp时的读法
	var resultMap map[string][]string
	err = json.Unmarshal([]byte(body), &resultMap)

	file, _ := os.OpenFile("getWrongTrace.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file.Close()
	fmt.Fprintln(file,string(body))

	if err != nil {
		log.Fatalf("getWrongTrace json unmarshal err=%v,str=%v\n", err, string(body))
	}
	return resultMap
}
