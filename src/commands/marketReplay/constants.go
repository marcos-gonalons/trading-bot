package main

type Side string

const (
	LONGS_SIDE  Side = "longs"
	SHORTS_SIDE Side = "shorts"
)

type ReplayType string

const (
	SINGLE_TYPE ReplayType = "single"
	COMBO_TYPE  ReplayType = "combo"
)
