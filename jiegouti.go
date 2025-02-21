package main

import "fmt"

//定义结构体
type kou1 struct {
	id   int
	name string
	xb   bool
	diqu string
}

func main() {
	var wangshuo kou1
	fmt.Println(wangshuo)
}
