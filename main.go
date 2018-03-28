package main

import (
	"github.com/shadowsocks/go-shadowsocks2/core"
	"net/url"
	"os"
	"syscall"
	"fmt"
	"os/signal"
	"strings"
	mconf "villeSock/src/config"
	vllog "common/villog"
	"villeSock/src/handle"
)

/**
 * ciph ：Cipher interface类型 主要包含 StreamConn(net.Conn) net.Conn
 * 和 PacketConn(net.PacketConn) net.PacketConn
 */
func runWork(user *mconf.UserGroup)( err error){

	addr := user.Server
	cipher := user.Cipher
	password := user.Password
	//判断是否以 ss://开头 (ss://是 websocket连接协议)
	if strings.HasPrefix(addr, "ss://") {
		addr, cipher, password, err = parseURL(addr)
		if err != nil {
			vllog.LogE("error :",err)
		}
	}else{
		addr = fmt.Sprintf("%s:%d",user.Server, user.Port)
	}
	ciph, err := core.PickCipher(cipher, []byte(user.Key), password)
	if err != nil {
		vllog.LogE("Error :", err)
		return err
	}
	vllog.LogI("addr = %s",addr)
	go handle.UdpRemote(addr, user.UDPTimeout ,ciph.PacketConn)
	go handle.TcpRemote(addr, ciph.StreamConn)
	return nil
}

func main() {
	//获取参数
	confArgs := mconf.ArgsPare()
	for _, user := range confArgs.UserGroups{
		if err := runWork(user); err != nil{
			os.Exit(1)
		}
	}

	//检测系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	processStr := <-sigCh
	fmt.Printf("退出 shadowsocket-go 进程 %v", processStr );
}

/**
 * @author：注释-凌霄
 * 解析 url 参数 参数格式为：scheme://[userinfo@]host/path[?query][#fragment]
 * @return addr: host+post ,
 * @return cipher: 用户名
 * @return password:密码
 * @return err 错误信息
 */
func parseURL(s string) (addr, cipher, password string, err error) {
	/**
		URL类型代表一个解析后的URL（或者说，一个URL参照）。
		URL基本格式如下：
		scheme://[userinfo@]host/path[?query][#fragment]
		例如："postgres://user:pass@host.com:5432/path?k=v#f"
		Host 同时包括主机名和端口信息，如过端口存在的话，使用 strings.Split() 从 Host 中手动提取端口
		User 包含了所有的认证信息，这里调用 Username和 Password 来获取独立值
	 */
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	addr = u.Host
	if u.User != nil {
		cipher = u.User.Username()
		password, _ = u.User.Password()
	}
	return
}
