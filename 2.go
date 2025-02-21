package main

import (
	"fmt"
	"strconv"
)

func add1(pk int, users map[string]map[string]string) {
	var (
		id   string = strconv.Itoa(pk) // 由int转为字符串类型
		name string
		age  string
	)
	//fmt.Println("请输入ID：")
	//fmt.Scanln(&id)
	fmt.Println("请输入名字：")
	fmt.Scanln(&name)
	fmt.Println("请输入职位：")
	fmt.Scanln(&age)
	users[id] = map[string]string{
		"id":   id,
		"name": name,
		"age":  age,
	}
}
func main() {

	//asdh := func(name string) {
	//	fmt.Println("hello ", name)
	//}
	//asdh("kou")
	fmt.Println("欢迎使用商汤系统")
	users := make(map[string]map[string]string) // 初始化用户数据存储
	var pk int
	for {
		var tang string
		fmt.Print(`
		1.新建用户
		2.修改
		3.删除
		4.查询
		5.退出
		请输入指令：
		`)
		// 从用户输入中读取值
		fmt.Scanln(&tang)

		if tang == "1" {
			fmt.Println("新建用户功能")
			add1(pk, users)
			pk++

		} else if tang == "2" {
			fmt.Println("修改功能")
		} else if tang == "3" {
			fmt.Println("删除功能")
		} else if tang == "4" {
			fmt.Println("查询功能")
			fmt.Println(users) // 打印当前用户信息
		} else if tang == "5" {
			fmt.Println("退出系统")
			break
		} else {
			fmt.Println("指令输入错误，请检查指令")
		}
	}
}
