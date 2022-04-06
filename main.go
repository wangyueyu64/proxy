package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Pxy struct{}

func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger.Infof("Received request %s %s \n", req.Method, req.Host)

	if filter(req) == false {
		http.Error(rw, "安全评估结果不满足访问要求", http.StatusBadRequest)
		return
	}

	client := http.Client{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", "localhost:8081", req.RequestURI)

	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadGateway)
		return
	}

	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)

	defer resp.Body.Close()
}

type Computer struct {
	ComputerId string `json:"computer_id"`
	Mac        string `json:"mac"`
	Model      string `json:"model"`
	Os         string `json:"os"`
	User       string `json:"user"`
	Level      int    `json:"level"`
}

func filter(req *http.Request) bool {

	var usrLevel int
	var sourceLevel int
	var opLevel int
	var score float64
	////*********操作所需安全等级
	//var opFind int = 90
	//var opAdd int = 80
	//var opUpdate int = 70
	//var opDel int = 60

	url := strings.Split(req.URL.String(), "?")[0]

	//过滤请求url
	if url == "/user/sendEmail" || url == "/user/login" {
		return true
	}

	if url != "/computer/add" && url != "/computer/find" && url != "/computer/update" && url != "/computer/del" {
		return false
	}

	//********操作安全级别
	if url == "/computer/find" {
		opLevel = 90
	}
	if url == "/computer/add" {
		opLevel = 80
	}
	if url == "/computer/update" {
		opLevel = 70
	}
	if url == "/computer/del" {
		opLevel = 60
	}

	//********获取用户安全等级
	userid := req.Header.Get("userid")
	resp, _ := http.Get("http://localhost:8081/user/getLevel?userid=" + userid)
	defer resp.Body.Close()

	buf := make([]byte, 1024)
	for {
		// 接收服务端信息
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			logger.Errorf("read body err: %v", err)
			return false
		} else {
			res := string(buf[:n])
			usrLevel, err = strconv.Atoi(res)
			if err != nil {
				logger.Errorf("Atoi err: %v", err)
				return false
			}
			break
		}
	}

	//*******获取资源安全等级
	//不存在 使用资源默认等级
	if url == "/computer/add" {
		sourceLevel = 5
	} else {
		//已有资源根据id 查询
		var computerId string

		if url == "/computer/del" {
			//获取computerId
			computerId = strings.Split(req.URL.String(), "id=")[1]
		} else {
			computerId = strings.Split(strings.Split(req.URL.String(), "id=")[1], "&")[0]
		}

		resp1, _ := http.Get("http://localhost:8081/computer/find?id=" + computerId)
		defer resp1.Body.Close()

		var res []Computer

		n, err := ioutil.ReadAll(resp1.Body)
		if err != nil && err != io.EOF {
			logger.Errorf("read body err: %v", err)
			return false
		} else {
			fmt.Printf("n :%v\n", string(n))
			if err = json.Unmarshal([]byte(string(n)), &res); err != nil {
				logger.Errorf("Unmarshal err: %v", err)
				return false
			}
		}

		fmt.Printf("res :%v\n", res)
		sourceLevel = res[0].Level
	}

	score = float64(usrLevel * opLevel / sourceLevel)
	fmt.Printf("score :%v\n", score)

	return true
}

func main() {
	if err := LogInit(); err != nil {
		fmt.Printf("init log failed err :%v\n", err)
	}
	fmt.Println("Serve on :8080")
	http.Handle("/", &Pxy{})
	http.ListenAndServe(":8080", nil)
}
