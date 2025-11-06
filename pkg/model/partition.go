package model

type Partitions []Partition

type Partition struct {
	Name string
	Uuid string
	Size int64
}
