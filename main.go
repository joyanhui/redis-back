package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	ctx     = context.Background()
	redisDb redis.UniversalClient
)

func main() {
	args := os.Args
	//log.Println(args)
	if len(args) < 3 {
		log.Println("参数错误 正确格式应该是 : redis-back load/dump date.json redis://用户名:密码@任意节点地址:10601")
		return
	}
	log.Println("===== redis连接 初始化... =====  ")
	redisDb = GetRedisClient(args)
	if redisDb == nil {
		log.Println("===== 初始化错误 进程停止... =====  ")
		return
	}
	switch args[1] {
	case "load":
		log.Println("加载json文件 to redis_serv", args[2], args[3])
		log.Println("===== 正在导入...=====")
		load_from_json(args[2], redisDb)
	case "dump":
		log.Println("导出json文件 form redis_serv", args[2], args[3])
		log.Println("===== 正在导出...=====")
		dump(args[2], redisDb)
	}
}
