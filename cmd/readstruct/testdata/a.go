package main

type A struct {
	AA int    `key:"val"`
	bb string `key:"val"`
	CC struct {
		x int
	}
}

type B struct {
	*A
	BB string `key2:"val"`
}

func (b B) Method1() {
}

func (b *B) Method2(str string) (ret int) {
	return
}

func main() {
}
