import base64
import re
import requests
from bs4 import BeautifulSoup
from tqdm import tqdm

def fofa():
    # 获取用户输入的搜索数据
    search_data = input("请输入关键词：")

    # 获取用户输入的爬取的结束页数
    end_page = input("请输入爬取的页数：")

    # 确保输入的页数是整数
    try:
        end_page = int(end_page)
    except ValueError:
        print("页数必须是整数。")
        exit()

    search_data_bs=base64.b64encode(search_data.encode('utf-8')).decode('utf-8')
    url='https://fofa.info/result?qbase64='
    headers={
        'cookie':'refresh_token=1;'
                 'fofa_token=eyJhbGciOiJIUzUxMiIsImtpZCI6Ik5XWTVZakF4TVRkalltSTJNRFZsWXpRM05EWXdaakF3TURVMlkyWTNZemd3TUdRd1pUTmpZUT09IiwidHlwIjoiSldUIn0.eyJpZCI6NDU0NTkyLCJtaWQiOjEwMDI2MTY4OSwidXNlcm5hbWUiOiJzbnNiY2NiY2RoaGRzc2g4NzcyIiwiZXhwIjoxNzEwODI4NTA1fQ.J-0AWmvjfRlEVm23ac4FNHwZiLSm7O6k6fU3WDUEckUpKPCiP3_y5CxSn34tw35KyRmT0sbDxdfYjtLAPy5Puw;'
    }

    with open('ip.txt', 'w') as f:
        # 从第一页开始爬取，直到结束页
        for yeshu in tqdm(range(1, end_page + 1)):
            urls=url+search_data_bs+"&page="+str(yeshu)

            response = requests.get(urls, headers=headers)
            soup = BeautifulSoup(response.text, 'html.parser')

            # 提取所有的IP:port
            ip_ports = re.findall(r'\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}:[0-9]{1,5}\b', soup.text)
            for ip_port in ip_ports:
                f.write(ip_port + '\n')

fofa()