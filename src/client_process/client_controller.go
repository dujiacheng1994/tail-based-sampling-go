package client_process

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func init() {
	//log
	file, _ := os.OpenFile("../log/client.log", os.O_RDWR|os.O_CREATE, 0666) //打开日志文件，不存在则创建
	log.SetOutput(file) //设置输出流
	log.SetFlags(log.Lshortfile)

	http.HandleFunc("/getWrongTrace", GetWrongTraceHandler)
}

func GetWrongTraceHandler(w http.ResponseWriter, req *http.Request) {
	traceIdListStr := req.FormValue("traceIdList")
	batchPosStr := req.FormValue("batchPos")
	batchPos, err := strconv.Atoi(batchPosStr)
	if err != nil {
		fmt.Println("GetWrongTraceHandler.atoi.err:,str:", err, batchPosStr)
	}
	jsonStr := getWrongTracing(traceIdListStr, batchPos)
	//fmt.Printf("suc to getWrongTrace, batchPos:%v,jsonStr:%v\n", batchPos, jsonStr)
	w.Write([]byte(jsonStr))
}
