package main

import "fmt"



func myfunc(){
	i := 0
	HERE:
	fmt.Println(i)
	i++
	if i < 10 {
		goto HERE
	}
}

//不定量参数

func Myfunc2(args ...int){
	for _,arg := range args{
		fmt.Print(arg,",")
	}
}

//不定变量中的不同类型

func MyPrintf(args ...interface{}){



	for _,arg := range args {
		switch arg.(type) {
		case int:
			fmt.Println(arg," is an int value")
		case string:
			fmt.Println(arg," is an string value")
		case int64:
			fmt.Println(arg," is an int64 value")
		default:
			fmt.Println(arg," is an unknown value")
			
		}
	}
}


func main(){

	var v1 int =1
	var v2 int64 = 234
	var v3 string = "hello"
	var v4 float32 = 1.23



	var myarray [10]int = [10]int{1,2,3,4,5,6,7,8}

	var myslice []int = myarray[5:]
	myslice2 := make([]int,5,10)

	fmt.Println("Element of myarray:")
	for _,v := range myarray{
		fmt.Print(v,",")
	}

	fmt.Println("\nElement of myslice:")
	for _,v := range myslice{
		fmt.Print(v,",")
	}

	fmt.Println("\nlen(myslice2)",len(myslice2))
	fmt.Println("\ncap(myslice2)",cap(myslice2))

	myslice = append(myslice,1,2,3)
	myslice = append(myslice,myslice2...)   //append函数的变量是个数组切片的话，要加省略号

	fmt.Println("\nElement of myslice:")
	for _,v := range myslice{
		fmt.Print(v,",")
	}

	fmt.Print()
	myfunc()
	Myfunc2(1,2,3)
	MyPrintf(v1,v2,v3,v4)

	//hello.Sayhello()
}
