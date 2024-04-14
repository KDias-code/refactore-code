package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string
	fT         string
	taskResult []byte
}

func main() {
	taskCreator := func(a chan Ttype) {
		go func() {
			for {
				ft := time.Now().Format(time.RFC3339)
				if time.Now().Nanosecond()%2 > 0 {
					ft = "Some error occurred"
				}
				a <- Ttype{cT: ft, id: int(time.Now().Unix())}
				time.Sleep(time.Millisecond * 500)
			}
		}()
	}

	taskWorker := func(t Ttype) Ttype {
		tt, _ := time.Parse(time.RFC3339, t.cT)
		if tt.After(time.Now().Add(-20 * time.Second)) {
			t.taskResult = []byte("task has been succeeded")
		} else {
			t.taskResult = []byte("something went wrong")
		}
		t.fT = time.Now().Format(time.RFC3339Nano)

		time.Sleep(time.Millisecond * 150)

		return t
	}

	taskSorter := func(t Ttype, wg *sync.WaitGroup, result map[int]Ttype, errChan chan error) {
		defer wg.Done()
		if string(t.taskResult[14:]) == "succeeded" {
			result[t.id] = t
		} else {
			errChan <- fmt.Errorf("task id %d time %s, error %s", t.id, t.cT, t.taskResult)
		}
	}

	superChan := make(chan Ttype, 10)
	doneTasks := make(chan Ttype)
	errChan := make(chan error)
	result := make(map[int]Ttype)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range superChan {
			go func(t Ttype) {
				t = taskWorker(t)
				taskSorter(t, &wg, result, errChan)
			}(t)
		}
	}()

	go taskCreator(superChan)

	go func() {
		wg.Wait()
		close(doneTasks)
		close(errChan)
	}()

	for {
		select {
		case t, ok := <-doneTasks:
			if !ok {
				doneTasks = nil
			} else {
				fmt.Println("Done task:", t.id)
			}
		case err := <-errChan:
			fmt.Println("Error:", err)
		default:
			time.Sleep(100 * time.Millisecond)
		}

		if doneTasks == nil && len(errChan) == 0 {
			break
		}
	}
}
