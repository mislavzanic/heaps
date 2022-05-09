package Brodal

type PriorityQueue interface {
	Meld(PriorityQueue)
	Min() float64
	DeleteMin()
	Insert(float64)
}
