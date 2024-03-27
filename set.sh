sudo apt update -y
sudo apt upgrade -y
sudo apt install -y masscan
sudo apt-get install -y libpcap-dev
wget https://raw.githubusercontent.com/robertdavidgraham/masscan/master/data/exclude.conf
pip install requests
pip install tqdm
pip install beautifulsoup4
go mod init test
go get github.com/vbauerster/mpb/v7
echo 注意!使用ip段扫描可能较慢,非大型扫描不建议加--excludefile exclude.conf
echo 以下为示例语法:
echo sudo masscan 18.180.0.0/15 -p1-65535 --rate 10000 -oX scan.xml --excludefile exclude.conf