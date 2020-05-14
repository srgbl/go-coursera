package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for d := range in {
		wg.Add(1)
		go func(d interface{}) {
			wg2 := &sync.WaitGroup{}
			wg2.Add(2)

			data := strconv.Itoa(d.(int))

			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()

			r := make([]string, 2)
			go func() {
				r[0] = DataSignerCrc32(data)
				wg2.Done()
			}()
			go func() {
				r[1] = DataSignerCrc32(md5)
				wg2.Done()
			}()

			wg2.Wait()
			out <- strings.Join(r, "~")
			wg.Done()
		}(d)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for d := range in {
		wg.Add(1)
		go func(d interface{}) {
			wg2 := &sync.WaitGroup{}

			data := d.(string)
			res := make([]string, 6)
			for i := 0; i < 6; i++ {
				wg2.Add(1)
				go func(ind int, d string) {
					res[ind] = DataSignerCrc32(strconv.Itoa(ind) + d)
					wg2.Done()
				}(i, data)
			}

			wg2.Wait()
			out <- strings.Join(res, "")
			wg.Done()
		}(d)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var res []string
	for d := range in {
		data := d.(string)
		res = append(res, data)
	}
	sort.Strings(res)
	out <- strings.Join(res, "_")
}

func ExecutePipeline(jj ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})

	for _, j := range jj {
		wg.Add(1)
		go func(jo job, i, o chan interface{}) {
			defer wg.Done()
			defer close(o)
			jo(i, o)
		}(j, in, out)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

/*
func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	// inputData := []int{0,1}

	var testResult string
	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, _ := dataRaw.(string)
			testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)

	end := time.Since(start)
	fmt.Println(end)
	fmt.Println(testResult)
}
*/
