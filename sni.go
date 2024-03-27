package main

import (
	"bufio"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

func checkSNI(ip string) string {
	dialer := &net.Dialer{
		Timeout: 60 * time.Second,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", ip+":443", &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "netflix.com",
	})
	if err != nil {
		return ""
	}
	defer conn.Close()

	err = conn.Handshake()
	if err != nil {
		return ""
	}

	cert := conn.ConnectionState().PeerCertificates[0]
	if strings.Contains(cert.Subject.String(), "netflix.com") {
		return ip
	}

	return ""
}

func checkIP(ip string) string {
	resp, err := httpClient.Get("http://" + ip + ":80")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	resp, err = httpClient.Get("http://103.75.70.133:80")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	targetBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	if string(body) == string(targetBody) {
		return ip
	}

	return ""
}

func main() {
	file, err := os.Open("123.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ips := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
		ip := re.FindString(line)
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	tempFile, err := ioutil.TempFile("", "sni")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	for _, ip := range ips {
		_, err := tempFile.WriteString(ip + "\n")
		if err != nil {
			panic(err)
		}
	}

	tempFile.Seek(0, 0) // Reset the file pointer to the beginning of the file

	scanner = bufio.NewScanner(tempFile)
	ips = []string{}
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	file, err = os.Create("sni_success.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	p := mpb.New(mpb.WithWaitGroup(&sync.WaitGroup{}))
	bar := p.AddBar(int64(len(ips)),
		mpb.PrependDecorators(
			decor.CountersNoUnit("%d / %d", decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
		),
	)

	var wg sync.WaitGroup
	successIPs := make(chan string, len(ips))

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			if checkSNI(ip) != "" && checkIP(ip) != "" {
				successIPs <- ip
			}
			bar.Increment()
		}(ip)
	}

	go func() {
		wg.Wait()
		close(successIPs)
	}()

	for ip := range successIPs {
		_, err := file.WriteString(ip + "\n")
		if err != nil {
			panic(err)
		}
		file.Sync() // Flush the data to disk immediately after writing
	}

	p.Wait()
}
