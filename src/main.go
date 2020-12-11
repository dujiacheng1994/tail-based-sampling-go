package main

import (
	"dujiacheng.jason/tail-based-sampling-go/src/backend_process"
	"dujiacheng.jason/tail-based-sampling-go/src/client_process"
	"dujiacheng.jason/tail-based-sampling-go/src/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func init(){
}

func main() {
	// common controller
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/setParameter", setParameterHandler)
	http.HandleFunc("/start", startHandler)

	// 分port的服务
	port := getPort()
	if port == 8002 {
		backend_process.Start()
	}
	fmt.Println("http://localhost:"+strconv.Itoa(port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

func readyHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write([]byte("suc"))
}

func setParameterHandler(w http.ResponseWriter, req *http.Request) {
	key := "port"
	dataPortStr := req.URL.Query().Get(key)
	resp := ""
	port := getPort()
	utils.DataSourcePort = dataPortStr
	fmt.Println("set dataPort:",utils.DataSourcePort)
	if port == 8000 || port == 8001 {
		go client_process.Start()
	}
	resp = "suc"
	_, _ = w.Write([]byte(resp))
}

func startHandler(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write([]byte("suc"))
}

func getPort() int {
	args := os.Args
	for i , v := range args{
		fmt.Println("args"+strconv.Itoa(i)+":"+v)
	}
	portStr := args[1]
	fmt.Println("server port:", portStr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("wrong port arg,err=",err)
	}
	return port
}