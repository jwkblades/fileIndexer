package main

import "fmt"
import "flag"
import "path/filepath"
import "os"
import "sync"
import "container/list"
import "io/ioutil"
import "regexp"

func main() {
    flag.Parse()
    root := flag.Arg(0)
    threads := *flag.Int("t", 1, "The number of threads to run")

    var wg sync.WaitGroup
    wordMap := make(map[string]uint)
    wordMapLock := &sync.Mutex{}
    fsCrawlCompleted := false
    fileQueue := list.New()
    fileQueueLock := &sync.Mutex{}

    wg.Add(threads)
    for i := 0; i < threads; i++ {
        go func() {
            //fmt.Println("Looking at entry ... in thread...")
            threadWordMap := make(map[string]uint)
            var path *list.Element = nil
            var filePath string = ""
            var word string = ""

            //fmt.Println("fsCrawlCompleted: ", fsCrawlCompleted, ", queue length: ", fileQueue.Len())
            for !fsCrawlCompleted || fileQueue.Len() > 0 {
                fmt.Println("LOOPING fsCrawlCompleted: ", fsCrawlCompleted, ", queue length: ", fileQueue.Len())
                filePath = ""
                if fileQueue.Len() > 0 {
                    fileQueueLock.Lock()
                    path = nil
                    if fileQueue.Len() > 0 {
                        path = fileQueue.Front()
                        if path != nil {
                            filePath = path.Value.(string)
                            fileQueue.Remove(path)
                        }
                    }
                    fileQueueLock.Unlock()
                    //fmt.Println("Looking at provided file path: ", filePath)
                }

                if filePath != "" {
                    fmt.Println("Found path to read through: ", path)
                    contents, err := ioutil.ReadFile(filePath)
                    if err != nil {
                        continue
                    }

                    var startNewWord bool = false
                    for i := range contents {
                        var v byte = contents[i]
                        //fmt.Println("Character: ", i, v)

                        if startNewWord {
                            word = ""
                            startNewWord = false
                        }

                        if (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') || (v >= '0' && v <= '9') {
                            word += string([]byte{v})
                        } else {
                            startNewWord = true
                        }

                        if word == "" || !startNewWord {
                            continue
                        }

                        //fmt.Println("Adding word: ", word)
                        threadWordMap[word]++
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

    r, _ := regexp.Compile("\\.txt$")
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        //fmt.Printf("Visited: %s\n", path)
        if !info.Mode().IsDir() && (info.Mode() & os.ModeSymlink) == 0 && r.MatchString(path) {
            fileQueue.PushBack(path)
            //fmt.Println("Added path: ", path, " to queue")
        }
        return nil
    })
    fmt.Printf("Filesystem walk returned %v\n", err);
    fsCrawlCompleted = true

    wg.Wait()

    for k, v := range wordMap {
        fmt.Println(k, v)
    }
}

