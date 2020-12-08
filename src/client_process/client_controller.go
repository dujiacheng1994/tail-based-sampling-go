package client_process

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func init() {
	log.SetPrefix("TRACE: ")
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
