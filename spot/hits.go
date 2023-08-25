package main

import "image"

type Hit[T any] struct {
	R     image.Rectangle
	Value T
}

func (h *Hit[T]) Contains(x, y int) bool {
	return image.Pt(x, y).In(h.R)
}

type Hits[T any] []Hit[T]

func (hits Hits[T]) Find(x, y int) *Hit[T] {
	for _, h := range hits {
		if h.Contains(x, y) {
			return &h
		}
	}

	return nil
}
