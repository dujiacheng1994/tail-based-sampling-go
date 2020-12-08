package client_process

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BATCH_SIZE  = 20000
	BATCH_COUNT = 15
)

var (
	batchTraceList         []*sync.Map
	batchTraceListNotEmpty []bool
	mu                     sync.Mutex
)

func init() {
	file, _ := os.OpenFile("op.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file.Close()
	log.SetOutput(file) //设置输出流
	log.SetPrefix("TRACE: ")
	batchTraceListNotEmpty = make([]bool, BATCH_COUNT)
	for i := 0; i < BATCH_COUNT; i++ {
		batchTraceList = append(batchTraceList, &sync.Map{})
	}
}

func Start() {
	go processData()
}

func processData() {
	path := getPath()
	resp, _ := http.Get(path)
	var (
		count int64
		pos   int32
	)
	badTraceIdList := make(map[string]bool)
	traceMap := batchTraceList[pos]

	// 测试：替代判断map len逻辑
	mu.Lock()
	batchTraceListNotEmpty[pos] = true
	mu.Unlock()
	br := bufio.NewReader(resp.Body)
	for {
		//lineStr, err := br.ReadString('\n')  // 这个函数会把\n也包括进来！
		line,_,err := br.ReadLine()
		lineStr := string(line)
		if err == io.EOF {
			break
		}
		count++
		cols := strings.Split(lineStr, "|")
		if cols != nil && len(cols) > 1 {
			traceId := cols[0]
			mu.Lock()
			v, _ := traceMap.LoadOrStore(traceId, []string{}) // 注意，这里存进去的也是个值而不是引用，这样子无法对此进行修改！实际上[]string也是在并发修改的！可以用数据库事务理解
			spanList := v.([]string)
			spanList = append(spanList, lineStr)
			traceMap.Store(traceId, spanList) // 需要修改后，再把值存回去，单线程时可以用数组指针的形式直接改，但多线程不行，否则失去了用sync.Map的意义
			// 测试是否成功插入
			//span, _ := batchTraceList[pos].Load(traceId)
			//fmt.Println("traceId:", traceId, "spanList:", span)
			mu.Unlock() // 用于保证事务的xx性？不然spanList存进去时，可能这期间已经traceMap更新了，形成了覆盖写

			if len(cols) > 8 {
				tags := cols[8]
				if tags != "" {
					if strings.Contains(tags, "error=1") {
						badTraceIdList[traceId] = false
					} else if strings.Contains(tags, "http.status_code=") && !strings.Contains(tags, "http.status_code=200") {
						badTraceIdList[traceId] = false
					}
				}
			}
		}
		if count%BATCH_SIZE == 0 {
			fmt.Println(pos)
			pos++
			if pos >= BATCH_COUNT {
				pos = 0
			}
			traceMap = batchTraceList[pos]

			// TODO to use lock/notify
			if batchTraceListNotEmpty[pos] {
				fmt.Println("waiting for pos:", pos)
				for {
					time.Sleep(10 * time.Millisecond)
					if !batchTraceListNotEmpty[pos] {
						fmt.Println("pos ready:", pos)
						break
					}
				}
			}
			// TODO to use lock/notify
			// if len(traceMap) > 0 {
			//	for{
			//		time.Sleep(10 * time.Millisecond)
			//		if len(traceMap) == 0{
			//			break
			//		}
			//	}
			//}
			batchPos := count/BATCH_SIZE - 1
			updateWrongTraceId(badTraceIdList, int(batchPos))
			badTraceIdList = make(map[string]bool)
			fmt.Println("suc to updateBadTraceId, batchPos:", batchPos)

			// test
			//if pos == 2 {
			//	time.Sleep(time.Hour)
			//}
		}
	}
	updateWrongTraceId(badTraceIdList, (int)(count/BATCH_SIZE-1))
	callFinish()
}

func callFinish() {
	_, _ = http.Get("http://localhost:8002/finish")
}

func clearMap(list *sync.Map) {
	list.Range(func(key interface{}, value interface{}) bool {
		list.Delete(key)
		return true
	})
}

func updateWrongTraceId(badTraceIdList map[string]bool, batchPos int) {
	list := make([]string, 0, 12)
	for k, _ := range badTraceIdList {
		list = append(list, k)
	}
	jsonStr, _ := json.Marshal(list)
	fmt.Println("updateBadTraceId, json:" + string(jsonStr) + ", batch:" + strconv.Itoa(batchPos))

	_, err := http.PostForm("http://localhost:8002/setWrongTraceId", url.Values{"traceIdList": {string(jsonStr)}, "batchPos": {strconv.Itoa(batchPos)}})
	if err != nil {
		fmt.Println("updateBadTraceId err.", err)
	}
}

func getWrongTracing(traceIdListStr string, batchPos int) string {
	//fmt.Printf("getWrongTracing, batchPos:%d, wrongTraceIdList:\n %s\n", batchPos, traceIdListStr)
	var traceIdList []string
	err := json.Unmarshal([]byte(traceIdListStr), &traceIdList)
	if err != nil {
		log.Fatalf("getWrongTracing json unmarshal err.%v", err)
	}
	wrongTraceMap := make(map[string][]string)
	pos := batchPos % BATCH_COUNT
	previous := pos - 1
	if previous == -1 {
		previous = BATCH_COUNT - 1
	}
	next := pos + 1
	if next == BATCH_COUNT {
		next = 0
	}
	getWrongTraceWithBatch(previous, pos, traceIdList, wrongTraceMap)
	getWrongTraceWithBatch(pos, pos, traceIdList, wrongTraceMap)
	getWrongTraceWithBatch(next, pos, traceIdList, wrongTraceMap)

	// to clear spans, don't block client process thread. TODO to use lock/notify
	mu.Lock()
	clearMap(batchTraceList[previous])
	batchTraceListNotEmpty[previous] = false
	mu.Unlock()

	wrongTraceMapStr, err := json.Marshal(wrongTraceMap)
	if err != nil {
		log.Fatalf("getWrongTracing json marshal err.%v", err)
	}
	return string(wrongTraceMapStr)
}

func getWrongTraceWithBatch(batchPos, pos int, traceIdList []string, wrongTraceMap map[string][]string) {
	// donot lock traceMap,  traceMap may be clear anytime.
	traceMap := batchTraceList[batchPos]

	//test
	//fmt.Println("print traceMap", "pos:", batchPos)
	//fmt.Println(traceMap)
	//fmt.Println(traceMap.Load("44eddbffae8ee745"))
	//traceMap.Range(func(key, value interface{}) bool {
	//	log.Println(key, "=", value)
	//	fmt.Println(key, "=", value)
	//	return true
	//})

	for _, traceId := range traceIdList {
		v, ok := traceMap.Load(traceId)
		if !ok {
			//fmt.Println("fail to get wrongTraceMap from traceId:", traceId)
			continue
		}
		spanList, _ := v.([]string)
		//if spanList != nil {
		// one trace may cross to batch (e.g batch size 20000, span1 in line 19999, span2 in line 20001)
		existSpanList := wrongTraceMap[traceId]
		if existSpanList != nil {
			existSpanList = append(existSpanList, spanList...)
		} else {
			wrongTraceMap[traceId] = spanList
		}
		// output spanlist to check 纯粹为了检查！
		spanListStr := strings.Join(spanList, "\n")
		fmt.Printf("\ngetWrongTracing, batchPos:%d, pos:%d, traceId:%s, spanList:\n %s", batchPos, pos, traceId, spanListStr)
		//}
	}
}

func getPath() string {
	args := os.Args
	port, _ := strconv.Atoi(args[1])
	path := ""
	if port == 8000 {
		path = "http://localhost:8080/trace1small.data"
	} else if port == 8001 {
		path = "http://localhost:8080/trace2small.data"
	}
	fmt.Println("getPath:", path)
	return path
}
