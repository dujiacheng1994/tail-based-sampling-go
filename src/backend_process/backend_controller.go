package backend_process

import (
	"dujiacheng.jason/tail-based-sampling-go/src/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const BATCH_COUNT = 90

var (
	mu                 sync.Mutex
	finishProcessCount int
	currentBatchPos    int
	traceIdBatchList   []*TraceIdBatch   // 这里，如果要写入，就要用指针形式！！不然每次赋值到函数中，如traceIdBatch := traceIdBatchList[pos]时，会拷一个副本！而不是引用
)

func init() {
	//log
	file, _ := os.OpenFile("../log/server.log", os.O_RDWR|os.O_CREATE, 0666) //打开日志文件，不存在则创建
	log.SetOutput(file) //设置输出流
	log.SetFlags(log.Lshortfile)


	for i := 0; i < BATCH_COUNT; i++ {
		traceIdBatchList = append(traceIdBatchList, &TraceIdBatch{})
	}
	http.HandleFunc("/setWrongTraceId", setWrongTraceIdHandler)
	http.HandleFunc("/finish", finishHandler)
}

func setWrongTraceIdHandler(w http.ResponseWriter, req *http.Request) {
	traceIdListStr := req.FormValue("traceIdList")
	batchPosStr := req.FormValue("batchPos")
	batchPos, err := strconv.Atoi(batchPosStr)
	if err!= nil{
		fmt.Println("setWrongTraceIdHandler.atoi.err:,str:",err,batchPosStr)
	}
	pos := batchPos % BATCH_COUNT
	var traceIdList []string
	err = json.Unmarshal([]byte(traceIdListStr), &traceIdList)
	if err != nil {
		fmt.Println("SetWrongTraceIdHandler json unmarshal fail, str:", traceIdListStr)
	}

	// todo : 此处会并发写，要加锁！！
	mu.Lock()
	traceIdBatch := traceIdBatchList[pos]  // 注意，此处不像java引用拷贝，如果[]TraceIdBatch类型，就是值拷贝！所以不好
	if traceIdBatch.batchPos != 0 && traceIdBatch.batchPos != batchPos {
		fmt.Println("overwrite traceId batch when call setWrongTraceId")
	}
	traceIdBatch.batchPos = batchPos

	traceIdBatch.processCount = traceIdBatch.processCount + 1

	log.Printf("setWrongTraceId had called, batchPos:%v, traceIdList:%v,processcount:%v\n", batchPos, traceIdListStr,traceIdBatch.processCount)

	traceIdBatch.traceIdList = append(traceIdBatch.traceIdList, traceIdList...)

	mu.Unlock()

	w.Write([]byte("suc"))
}

func finishHandler(w http.ResponseWriter, req *http.Request) {
	// todo: 此处会并发写！要加锁
	mu.Lock()
	finishProcessCount = finishProcessCount + 1
	mu.Unlock()

	fmt.Println("receive call 'finish', count:", finishProcessCount)
	w.Write([]byte("suc"))
}

func isFinished() bool {
	for i := 0; i < BATCH_COUNT; i++ {
		currentBatch := traceIdBatchList[i]
		if currentBatch.batchPos != 0 {
			return false
		}
	}
	if finishProcessCount < utils.PROCESS_COUNT {
		return false
	}
	return true
}

func getFinishedBatch() *TraceIdBatch {
	next := currentBatchPos + 1
	if next >= BATCH_COUNT {
		next = 0
	}
	nextBatch := traceIdBatchList[next]
	currentBatch := traceIdBatchList[currentBatchPos]
	// when client process is finished, or then next trace batch is finished. to get checksum for wrong traces.
	if (finishProcessCount >= utils.PROCESS_COUNT && currentBatch.batchPos > 0) ||
		(nextBatch.processCount >= utils.PROCESS_COUNT && currentBatch.processCount >= utils.PROCESS_COUNT) {
		//reset
		newTraceIdBatch := TraceIdBatch{}
		traceIdBatchList[currentBatchPos] = &newTraceIdBatch
		currentBatchPos = next
		return currentBatch
	}
	return nil
}

//for{
//	if inited{
//		break
//	}
//	time.Sleep(100 * time.Millisecond)
//}
//
//for{
//	lock.Lock()
//	b := inited
//	lock.Unlock()
//	if b{
//		break
//	}
//	time.Sleep(100 * time.Millisecond)
//}
//
//func Setup2() <-chan bool {
//	time.Sleep(time.Second * 3)
//	c := make(chanbool)
//	c <- true
//	return c
//}
//
//func main() {
//	if <-Setup2(){
//		fmt.Println(“setup succeed”)
//	}
//}
