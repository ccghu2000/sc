package main

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func checkIP(ipPort string) string {
	resp, err := httpClient.Get("http://" + ipPort)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	if string(body) == targetBody {
		return ipPort
	}

	return ""
}

var targetBody string

func main() {
	file, err := os.Open("scan.xml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`<address addr="(\d+\.\d+\.\d+\.\d+)" addrtype="ipv4"/>.*?<port protocol="tcp" portid="(\d+)">`)
	ipList := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			ipList = append(ipList, matches[1]+":"+matches[2])
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	resp, err := httpClient.Get("http://52.197.35.156:9002")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	targetBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	targetBody = string(targetBodyBytes)

	file, err = os.Create("ip_success.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	p := mpb.New(mpb.WithWaitGroup(&sync.WaitGroup{}))
	bar := p.AddBar(int64(len(ipList)),
		mpb.PrependDecorators(
			decor.CountersNoUnit("%d / %d", decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
		),
	)

	var wg sync.WaitGroup
	successIPs := make(chan string, len(ipList))

	for _, ipPort := range ipList {
		wg.Add(1)
		go func(ipPort string) {
			defer wg.Done()
			if checkIP(ipPort) != "" {
				successIPs <- ipPort
			}
			bar.Increment()
		}(ipPort)
	}

	go func() {
		wg.Wait()
		close(successIPs)
	}()

	for ipPort := range successIPs {
		_, err := file.WriteString(ipPort + "\n")
		if err != nil {
			panic(err)
		}
	}

	p.Wait()
}
