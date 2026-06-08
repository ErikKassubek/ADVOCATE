package main

import (
"sync"
"testing"
)

func Test10(t *testing.T) {
c := make(chan int, 10)
wg := sync.WaitGroup{}

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
c <- 1 }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
<-c }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
c <- 1 }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
<-c }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
c <- 1 }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
<-c }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
c <- 1 }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
<-c }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
c <- 1 }
})

wg.Go(func() {
for i := 0; i < 1000000000; i++ {
<-c }
})

wg.Wait()
}