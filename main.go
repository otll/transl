package main

import (
	"fmt"
	"transl/baidutransl"
)

func main() {
	// 百度翻译
	fmt.Println(baidutransl.Transl("This is baidu translate"))
}
