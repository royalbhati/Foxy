package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/muesli/termenv"
	"github.com/pkg/errors"
	"github.com/royalbhati/foxy/metadata"
	"github.com/royalbhati/foxy/progress"
)

type Download struct {
	URL           string
	TargetPath    string
	TotalSections int
	DirPath       string
}

func main() {
	p := termenv.ColorProfile()

	startTime := time.Now()
	var bar progress.Bar
	bar.NewOption(0, 10)

	var url string
	fmt.Printf("%s\n", termenv.String(" 🖊️ Paste the url here").Bold().Foreground(p.Color("#FDD835")))
	fmt.Scanf("%s\n", &url)

	if url == "" {
		log.Fatal("URL can't be empty")
	}

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	d := Download{
		URL:           url,
		TargetPath:    "",
		TotalSections: 10,
		DirPath:       path,
	}

	if err := d.Do(p, &bar); err != nil {
		fmt.Printf("\n%s\n", termenv.String("An error occured while downloading the file").Foreground(p.Color("0")).Background(p.Color("#E88388")))
		// fmt.Printf("\n%s", err)
		log.Fatal(err)
	}

	fmt.Printf("\n\n ✅ Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) Do(p termenv.Profile, bar *progress.Bar) error {

	fmt.Printf("\n%s\n", termenv.String(" 🔍 CHECKING URL ").Bold().Foreground(p.Color("#FDD835")))

	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	fmt.Printf("\n%s\n", termenv.String(" 👍 URL OK ").Bold().Foreground(p.Color("#8BC34A")))
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	a := metadata.GetFileName(resp)
	d.TargetPath = a

	if err != nil {
		return err
	}
	var yesOrNo string
	fmt.Printf("\n%s %f %s\n", termenv.String(" 🗂  SIZE IS ").Bold().Foreground(p.Color("#FFB300")), (float64(size)*float64(0.001))/1000, "MB")
	fmt.Printf("\n%s\n", termenv.String(" 🚀 PROCEED DOWNLOADING (y/n)").Bold().Foreground(p.Color("#FFB300")))
	fmt.Scanf("%s", &yesOrNo)

	if yesOrNo == "n" {
		fmt.Printf("%s\n", termenv.String("EXITING").Bold().Foreground(p.Color("#FF7043")))
		os.Exit(2)
		return nil
	}

	var sections = make([][2]int, d.TotalSections)

	eachSize := size / d.TotalSections

	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < d.TotalSections-1 {
			sections[i][1] = sections[i][0] + eachSize
		} else {
			sections[i][1] = size - 1
		}
	}

	var wg sync.WaitGroup
	sum2 := 0

	for i, s := range sections {
		wg.Add(1)

		go func(i int, s [2]int, bar *progress.Bar) {
			defer wg.Done()
			err = d.downloadSection(i, s, &sum2, p, bar)
			if err != nil {
				panic(err)
			}
		}(i, s, bar)

	}
	wg.Wait()
	return d.mergeFiles(sections)
}
func (d Download) downloadSection(i int, c [2]int, sum *int, p termenv.Profile, bar *progress.Bar) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(d.DirPath, fmt.Sprintf("section-%v.tmp", i)), b, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("section-%v.tmp", i)
	*sum = *sum + 1
	bar.Play(int64(*sum))

	return nil
}

// Get a new http request
func (d Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.URL,
		nil,
	)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "TDM")
	return r, nil
}

func (d Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(path.Join(d.DirPath, d.TargetPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("section-%v.tmp", i)
		b, err := ioutil.ReadFile(path.Join(d.DirPath, tmpFileName))
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}

	}

	return nil

}
