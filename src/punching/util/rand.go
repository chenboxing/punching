package util

import (
	"math/rand"
	"time"
)

// GenerateRandomPairKey 获取4位随机匹配码
func GenerateRandomPairKey() string {
	//97～122 小写字母
	rndNums := GenerateRandomNumber(97, 122, 4)
	key := ""
	for _, num := range rndNums {
		key = key + string(byte(num))
	}
	return key
}

//生成count个[start,end)结束的不重复的随机数
func GenerateRandomNumber(start int, end int, count int) []int {
	// Check the range
	if end < start || (end-start) < count {
		return nil
	}
	// Slice to store the result
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start) + start
		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}
