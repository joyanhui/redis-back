package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

func GetRedisClient(args []string) redis.UniversalClient {
	url := strings.Replace(args[3], "rediss-tcp", "rediss", 1)
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Println("redis连接字符串解析错误", args[3], err)
		return nil
	}
	if opt.TLSConfig != nil && len(args) > 3 { //自定义证书
		if args[4] == "skip" { //跳过证书校验
			opt.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		} else {
			if len(args) < 5 {
				log.Println("自定义ca/证书必须包含domain参数")
				return nil
			}
			if strings.Contains(args[4], "#") { //证书方式
				client_cert := strings.Split(args[4], "#")
				tlsConfig := &tls.Config{
					InsecureSkipVerify: false,
					ServerName:         args[5],
				}
				cert, err := tls.LoadX509KeyPair(client_cert[0], client_cert[1])
				if err != nil {
					log.Println("证书解析失败", args[4])
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
				opt.TLSConfig = tlsConfig
			} else { //ca方式
				caCert, err := os.ReadFile(args[4])
				if err != nil {
					log.Println("无法打开ca证书", args[4], err)
					return nil
				}
				tlsConfig := &tls.Config{
					RootCAs:            x509.NewCertPool(),
					InsecureSkipVerify: false,
					ServerName:         args[5],
				}
				ok := tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
				if !ok {
					log.Println("ca证书解析失败", args[4])
					return nil
				}
				opt.TLSConfig = tlsConfig
			}
		}

	}

	rdb := redis.NewClient(opt)

	// 测试连接
	result, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Println("当前redis连接错误", err)
		return nil
	}
	log.Println("当前节点ping测试", result)

	val, err := rdb.ClusterNodes(ctx).Result() //检查是否是集群
	if err != nil {
		log.Println("可能不是连接到的cluster集群,后面做单节点处理:", err)
		return rdb
	} else { //从单个节点中获取集群
		log.Println("当前应该是连接到集群")
		lines := strings.Split(val, "\n")
		var addrs []string
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				add := strings.Split(fields[1], "@")[0]
				addrs = append(addrs, add) //节点地址
			}
		}
		log.Println("节点地址：", addrs)
		var rdbC *redis.ClusterClient
		if strings.HasPrefix(args[3], "rediss-tcp") {
			rdbC = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    addrs,
				Username: opt.Username,
				Password: opt.Password,
			})
		} else {
			rdbC = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:     addrs,
				Username:  opt.Username,
				Password:  opt.Password,
				TLSConfig: opt.TLSConfig,
			})
		}
		return rdbC
	}

}
