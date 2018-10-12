package main

import "fmt"
import "flag"
import "path/filepath"
import "os"
import "sync"
import "container/list"
import "io/ioutil"
import "regexp"
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

    r, _ := regexp.Compile("\\.txt$")
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if r.MatchString(path) {
            //fmt.Printf("Path: '%s'\n", path)
            if (info.Mode() & os.ModeSymlink) == 0 {
                fileQueueLock.Lock()
                fileQueue.PushBack(path)
                fileQueueLock.Unlock()
                //fmt.Println("Added path: '", path, "' to queue")
            }
        }
        return nil
    })
    //fmt.Printf("Filesystem walk returned %v\n", err);

    wg.Add(*threads)
    //fmt.Printf("Starting up %d threads\n", *threads)
    for i := 0; i < *threads; i++ {
        go func() {
            //fmt.Println("Looking at entry ... in thread...")
            threadWordMap := make(map[string]uint)
            var path *list.Element = nil
            var filePath string = ""
            var word string = ""

            for fileQueue.Len() > 0 {
                //fmt.Println("LOOPING fsCrawlCompleted: ", fsCrawlCompleted, ", queue length: ", fileQueue.Len())
                filePath = ""
                fileQueueLock.Lock()
                path = nil
                if fileQueue.Len() > 0 {
                    path = fileQueue.Front()
                    if path != nil && path.Value != nil {
                        filePath = path.Value.(string)
                    }
                    fileQueue.Remove(path)
                }
                fileQueueLock.Unlock()

                if filePath != "" {
                    //fmt.Println("Looking at file: ", filePath)
                    contents, err := ioutil.ReadFile(filePath)
                    if err != nil {
                        continue
                    }

                    for i := range contents {
                        var v byte = contents[i]
                        //fmt.Println("Character: ", i, v)

                        if (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') || (v >= '0' && v <= '9') {
                            word += string([]byte{v})
                        } else if word != "" {
                            word = strings.ToLower(word)
                            //fmt.Println("Adding word: ", word)
                            threadWordMap[word]++
                            word = ""
                        }
                    }

                    if word != "" {
                        word = strings.ToLower(word)
                        //fmt.Println("Adding word: ", word)
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
            wg.Done()
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
