package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func dump(filename string, db redis.UniversalClient) {
	var redis_db_date []byte
	if dbC, ok := db.(*redis.ClusterClient); ok { //  Redis 集群
		err := dbC.ForEachMaster(ctx, func(ctx context.Context, masterDB *redis.Client) error { //遍历主节点
			redis_db_dateThis := dump_node(*masterDB)
			var m1 = make(map[string]interface{}) //合并
			var m2 = make(map[string]interface{})

			if len(redis_db_date) > 0 {
				err := json.Unmarshal(redis_db_date, &m1)
				if err != nil {
					log.Println("redis_db_date", err)
				}
			}
			if len(redis_db_dateThis) > 0 {
				err := json.Unmarshal(redis_db_dateThis, &m2)
				if err != nil {
					log.Println("redis_db_dateThis", masterDB, err)
				}
			}
			if len(m2) > 0 { // 合并两个映射
				for k, v := range m2 {
					m1[k] = v
				}
			}
			mergedData, err := json.Marshal(m1) // 将合并后的映射重新编码为 JSON 字符串
			if err != nil {
				log.Println("mergedData", err)
			} else {
				redis_db_date = mergedData
			}
			masterDB.Close()
			return nil
		})
		if err != nil {
			log.Println("主节点遍历失败", err)
			return
		}
	} else {
		if db2, ok := db.(*redis.Client); ok {
			redis_db_date = dump_node(*db2)
		} else {
			log.Println("你的redis 不是单点/主从或Cluster集群")
			return
		}
	}

	var tmp_data interface{} // 创建一个空接口用于解析 JSON 数据
	errJson := json.Unmarshal(redis_db_date, &tmp_data)
	if errJson != nil {
		fmt.Println("redis_db_date   JSON Unmarshal err :", errJson)
	} else {
		formattedJSON, err := json.MarshalIndent(tmp_data, "", "  ") // 格式化 JSON 数据
		if err != nil {
			log.Println("redis_db_date   JSON MarshalIndent err:", err)
		} else {
			redis_db_date = formattedJSON
		}
	}
	if len(redis_db_date) < 10 {
		log.Println("redis_db_date 长度不到10个字节 所以没有写入到本地文件")
		return
	}
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		filename = strings.Replace(filename, matches[1], time.Now().Format(matches[1]), -1)
		filename = strings.Replace(filename, "{", "", -1)
		filename = strings.Replace(filename, "}", "", -1)

	}
	log.Println("写入到文件", filename)
	os.MkdirAll(filepath.Dir(filename), 0755)
	err := os.WriteFile(filename, redis_db_date, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	log.Println("导出成功：")
	fmt.Println(filename)

}

func dump_node(db redis.Client) (jsonData []byte) {
	keys := db.Keys(ctx, "*").Val() // 获取所有键
	//fmt.Println("keys", keys)
	result := make(map[string]interface{}) // 创建一个空的 map 用于存储结果
	for _, key := range keys {             // 遍历键并获取对应的值和类型
		dataType := db.Type(ctx, key).Val()
		switch dataType { // 根据不同的数据类型获取值
		case "string":
			value := db.Get(ctx, key).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": value,
			}
		case "list":
			values := db.LRange(ctx, key, 0, -1).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": values,
			}
		case "hash":
			values := db.HGetAll(ctx, key).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": values,
			}
		case "set":
			values := db.SMembers(ctx, key).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": values,
			}
		case "zset":
			values := db.ZRange(ctx, key, 0, -1).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": values,
			}
		case "bitmaps":
			value := db.GetBit(ctx, key, 0).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": value,
			}
		case "hyperloglogs":
			value := db.PFCount(ctx, key).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": value,
			}
		case "geospatial":
			values := db.GeoRadius(ctx, key, 0, 0, &redis.GeoRadiusQuery{Unit: "km"}).Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": values,
			}
		case "bitfield":
			value := db.BitField(ctx, key, "GET", "u4", "0").Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": value,
			}
		case "stream":
			streamValues := db.XRange(ctx, key, "-", "+").Val()
			result[key] = map[string]interface{}{
				"type":  dataType,
				"value": streamValues,
			}
		default:
			fmt.Printf("Unsupported data type for key %s\n", key)
		}
	}
	jsonData, err := json.MarshalIndent(result, "", "  ") // 将结果转换为 JSON 格式
	if err != nil {
		fmt.Println("Error marshalling data to JSON:", err)
		return
	}
	return
}
