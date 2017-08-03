package util_test

import (
	"fmt"
	"math/rand"
	"punching/util"
	"testing"
	"time"
)

func TestGenerateRandomPairKey(t *testing.T) {
	t1 := util.GenerateRandomPairKey()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(122-97) + 97

	t.Log(string(byte(num)))

	fmt.Println("fff")
	fmt.Println(t1)
	if len(t1) != 4 {
		t.Errorf("长度不对,%s", t1)
	}
}
