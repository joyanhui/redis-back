package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func load_from_json(filename string, db redis.UniversalClient) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	restore_redis_db(data, db)
	log.Println("===== 导出完成 请仔细检查有没有错误...=====")

}

func restore_redis_db(jsonData []byte, db redis.UniversalClient) {
	var importedData map[string]interface{} // 解析导出的 JSON 数据
	err := json.Unmarshal(jsonData, &importedData)
	if err != nil {
		fmt.Println("Error unmarshalling JSON data:", err)
		return
	}
	for key, value := range importedData { // 遍历导入的数据并根据类型将其存储到 Redis
		dataMap, ok := value.(map[string]interface{})
		if !ok {
			fmt.Printf("Invalid data format for key %s\n", key)
			continue
		}
		dataType, ok := dataMap["type"].(string)
		if !ok {
			fmt.Printf("Invalid data type for key %s\n", key)
			continue
		}
		dataValue, ok := dataMap["value"]
		if !ok {
			fmt.Printf("Invalid data value for key %s\n", key)
			continue
		}
		switch dataType {
		case "string":
			err := db.Set(ctx, key, dataValue.(string), 0).Err()
			if err != nil {
				fmt.Printf("Error setting value for key %s: %v\n", key, err)
			}
		case "list":
			err := db.RPush(ctx, key, dataValue.([]interface{})...).Err()
			if err != nil {
				fmt.Printf("Error pushing values to list key %s: %v\n", key, err)
			}
		case "hash":
			hashValues, ok := dataValue.(map[string]interface{})
			if !ok {
				fmt.Printf("Invalid hash values for key %s\n", key)
				continue
			}
			err := db.HMSet(ctx, key, hashValues).Err()
			if err != nil {
				fmt.Printf("Error setting hash values for key %s: %v\n", key, err)
			}
		case "set":
			err := db.SAdd(ctx, key, dataValue.([]interface{})...).Err()
			if err != nil {
				fmt.Printf("Error adding values to set key %s: %v\n", key, err)
			}
		case "zset":
			zsetValues, ok := dataValue.([]map[string]interface{})
			if !ok {
				fmt.Printf("Invalid sorted set values for key %s\n", key)
				continue
			}
			var zsetEntries []redis.Z
			for _, zsetValue := range zsetValues {
				scoreStr, ok := zsetValue["score"].(string)
				if !ok {
					fmt.Printf("Invalid sorted set entry score for key %s\n", key)
					continue
				}
				score, err := strconv.ParseFloat(scoreStr, 64)
				if err != nil {
					fmt.Printf("Error parsing score for key %s: %v\n", key, err)
					continue
				}
				member, ok := zsetValue["value"].(string)
				if !ok {
					fmt.Printf("Invalid sorted set entry member for key %s\n", key)
					continue
				}
				zsetEntries = append(zsetEntries, redis.Z{
					Score:  score,
					Member: member,
				})
			}
			err := db.ZAdd(ctx, key, zsetEntries...).Err()
			if err != nil {
				fmt.Printf("Error adding values to sorted set key %s: %v\n", key, err)
			}
		case "bitmaps":
			bitmapsValues, ok := dataValue.(int64)
			if !ok {
				fmt.Printf("Invalid bitmaps values for key %s\n", key)
				continue
			}
			err := db.SetBit(ctx, key, 0, int(bitmapsValues)).Err()
			if err != nil {
				fmt.Printf("Error setting bitmaps values for key %s: %v\n", key, err)
			}
		case "hyperloglogs":
			hyperloglogsValues, ok := dataValue.([]interface{})
			if !ok {
				fmt.Printf("Invalid hyperloglogs values for key %s\n", key)
				continue
			}
			err := db.PFAdd(ctx, key, hyperloglogsValues...).Err()
			if err != nil {
				fmt.Printf("Error setting hyperloglogs values for key %s: %v\n", key, err)
			}
		case "geospatial":
			geospatialValues, ok := dataValue.(map[string]interface{})
			if !ok {
				fmt.Printf("Invalid geospatial values for key %s\n", key)
				continue
			}
			var geoLocation []*redis.GeoLocation
			for member, coordinates := range geospatialValues {
				coordinateMap, ok := coordinates.(map[string]interface{})
				if !ok {
					fmt.Printf("Invalid geospatial coordinate for key %s\n", key)
					continue
				}
				longitudeStr, ok := coordinateMap["longitude"].(string)
				if !ok {
					fmt.Printf("Invalid geospatial longitude for key %s\n", key)
					continue
				}
				latitudeStr, ok := coordinateMap["latitude"].(string)
				if !ok {
					fmt.Printf("Invalid geospatial latitude for key %s\n", key)
					continue
				}
				longitude, err := strconv.ParseFloat(longitudeStr, 64)
				if err != nil {
					fmt.Printf("Error parsing longitude for key %s: %v\n", key, err)
					continue
				}
				latitude, err := strconv.ParseFloat(latitudeStr, 64)
				if err != nil {
					fmt.Printf("Error parsing latitude for key %s: %v\n", key, err)
					continue
				}
				geoLocation = append(geoLocation, &redis.GeoLocation{
					Name:      member,
					Longitude: longitude,
					Latitude:  latitude,
				})
			}
			err := db.GeoAdd(ctx, key, geoLocation...).Err()
			if err != nil {
				fmt.Printf("Error adding values to geospatial key %s: %v\n", key, err)
			}
		case "bitfield":
			bitfieldValues, ok := dataValue.(string)
			if !ok {
				fmt.Printf("Invalid bitfield values for key %s\n", key)
				continue
			}
			err := db.BitField(ctx, key, bitfieldValues).Err()
			if err != nil {
				fmt.Printf("Error setting bitfield values for key %s: %v\n", key, err)
			}
		case "stream":
			streamEntries, ok := dataValue.([]interface{})
			if !ok {
				fmt.Printf("Invalid stream entries for key %s\n", key)
				continue
			}
			db.Del(ctx, key) // 这里需要根据情况判断是否应该先删掉  对应的key  不然原来key内stream有更大的id大的数据的情况下无法写入更小id的数据
			for _, entry := range streamEntries {
				entryMap, ok := entry.(map[string]interface{})
				if !ok {
					fmt.Printf("Invalid stream entry for key %s\n", key)
					continue
				}
				streamID, ok := entryMap["ID"].(string)
				if !ok {
					fmt.Printf("Invalid stream ID for key %s\n", key)
					continue
				}
				streamValues, ok := entryMap["Values"].(map[string]interface{})
				if !ok {
					fmt.Printf("Invalid values for stream entry with ID %s\n", streamID)
					continue
				}
				err := db.XAdd(ctx, &redis.XAddArgs{
					Stream: key,
					ID:     streamID,
					Values: streamValues,
				}).Err()
				if err != nil {
					fmt.Printf("Error adding entry to stream key %s: %v\n", key, err)
				}
			}
		default:
			fmt.Printf("Unsupported data type for key %s\n", key)
		}
	}
}
