package main

import "fmt"
import "sayhello"


func modify(array [10] int){
	array[0] = 10
	fmt.Println("In modify(), array values :",array)
}

func main(){
	array := [10]int{1,2,3,4,5}

	modify(array)

	fmt.Println("In main(), array values:",array)

	sayhello.Sayhello()

}
