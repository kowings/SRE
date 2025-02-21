package main

import "fmt"

func main() {
	// var a [10]int
	// fmt.Printf("%T", a)
	// a = [10]int{0: 10, 5: 20, 9: 100}
	// fmt.Println(a)      // 定义数组的每个值
	// fmt.Println(len(a)) //查看数组的长度

	var a = []int{1, 2, 3, 4, 5} //定义切片
	asdh := func(name, string) {
		fmt.Println(hello, name)
	}
	asdh("kou")
	// 删除切片中的3
	copy(a[2:], a[3:])
	a = a[:len(a)-1]
	fmt.Println(a)
}
