package main

import "fmt"

func main(){
	var j int = 5
	a := func() (func()){
		var i int = 10
		return func() {
			fmt.Println("i,j:",i,j)
		}
	}()
	a()
	fmt.Println()
	j *= 2
	a()
}
