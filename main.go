package main

import "fmt"
import "flag"
import "path/filepath"
import "os"
import "sync"
import "container/list"
import "io/ioutil"
import "sort"
import "strings"

type wordMapEntry struct {
    first string
    second uint
}

type wordMapArray []wordMapEntry

func (arr wordMapArray) Len() int {
    return len(arr)
}

func (arr wordMapArray) Swap(i int, j int) {
    arr[i], arr[j] = arr[j], arr[i]
}

func (arr wordMapArray) Less(i int, j int) bool {
    return arr[i].second > arr[j].second
}

func main() {
    threads := flag.Int("t", 1, "The number of threads to run")
    flag.Parse()
    root := flag.Arg(0)

    var wg sync.WaitGroup
    wordMap := make(map[string]uint)
    wordMapLock := &sync.Mutex{}
    fileQueue := list.New()
    fileQueueLock := &sync.Mutex{}

    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if path[len(path)-4:] == ".txt" && (info.Mode() & os.ModeSymlink) == 0 {
            fileQueue.PushBack(path)
        }
        return nil
    })

    nextPath := func() string {
        fileQueueLock.Lock()
        defer fileQueueLock.Unlock()
        var filePath string = ""

        if fileQueue.Len() > 0 {
            path := fileQueue.Front()
            if path != nil && path.Value != nil {
                filePath = path.Value.(string)
            }
            fileQueue.Remove(path)
        }

        return filePath
    }

    wg.Add(*threads)
    for i := 0; i < *threads; i++ {
        go func() {
            defer wg.Done()
            threadWordMap := make(map[string]uint)
            var filePath string = ""
            var word string = ""

            for fileQueue.Len() > 0 {
                filePath = nextPath()

                if filePath != "" {
                    contents, err := ioutil.ReadFile(filePath)
                    if err != nil {
                        continue
                    }
                    contentString := strings.ToLower(string(contents))

                    for i := range contentString {
                        var v byte = contentString[i]

                        if (v >= 'a' && v <= 'z') || (v >= '0' && v <= '9') {
                            word += string([]byte{v})
                        } else if word != "" {
                            threadWordMap[word]++
                            word = ""
                        }
                    }

                    if word != "" {
                        threadWordMap[word]++
                        word = ""
                    }
                }
            }

            wordMapLock.Lock()
            defer wordMapLock.Unlock()
            for k, v := range threadWordMap {
                wordMap[k] += v
            }
        }()
    }

    wg.Wait()

    var parts wordMapArray
    for k, v := range wordMap {
        parts = append(parts, wordMapEntry{k, v})
    }

    sort.Sort(parts)

    var max int = 10
    if len(parts) < max {
        max = len(parts)
    }
    for i := 0; i < max; i++ {
        fmt.Printf("%s\t\t%d\n", parts[i].first, parts[i].second)
    }
}
