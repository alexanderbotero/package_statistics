package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// struct used to store the name and count of the package
type PackageCount struct {
	Name  string
	Count int
}

// regular expression to validate architecture strings
// to avoid any attempts of attack via malformed URLs
var validArchRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// avoid magic number
const TopPackages = 10

// use the mirrors proposed
var defaultMirrors = []string{
	"http://deb.debian.org/debian/dists/stable/main/",
	"http://ftp.gr.debian.org/debian/dists/stable/main/",
	"http://ftp.us.debian.org/debian/dists/stable/main/",
}

func main() {
	// validate the architecture parameter is sent
	if len(os.Args) < 2 {
		// the architecure is not sent, explain how to user the cli
		log.Fatalf("Usage: %s <architecture>", os.Args[0])
	}
	arch := os.Args[1]

	// validate the arch parameter
	if !isValidArchitecture(arch) {
		// the parameter is not valid, explain to the user the expected format
		log.Fatalf("Error: invalid architecture format '%s'. Only lowercase letters, numbers, and hyphens are allowed.", arch)
	}

	// put a timeout, avoid hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// download the contents
	body, err := downloadContents(ctx, arch)
	if err != nil {
		// opps, something when wrong
		log.Fatalf("Error: %v", err)
	}
	defer body.Close()

	// unzip the content
	gzReader, err := gzip.NewReader(body)
	if err != nil {
		// opps, there is something wrong with the zip file
		log.Fatalf("Error uncompressing gzip: %v", err)
	}
	defer gzReader.Close()

	// parse the contents, get a map, where the name is the key,
	// and the value is amount of times that it was found
	packages, err := parseContents(gzReader)
	if err != nil {
		// opps, something was not right
		log.Fatalf("Error processing content: %v", err)
	}

	// get the top packages, avoid hardcoding the number
	// maybe a little over the top, but it will make it easy to change and test
	topN := calculateTopN(packages, TopPackages)
	// print the top packages
	printTopN(topN)
}

func isValidArchitecture(arch string) bool {
	// avoid funky characters
	return validArchRegex.MatchString(arch)
}

func downloadContents(ctx context.Context, arch string) (io.ReadCloser, error) {
	client := &http.Client{}

	// make a copy of the mirrors
	mirrors := make([]string, len(defaultMirrors))
	copy(mirrors, defaultMirrors)

	// randomize the mirrors
	// this is done to avoid that always the same server responds
	rand.Shuffle(len(mirrors), func(i, j int) {
		mirrors[i], mirrors[j] = mirrors[j], mirrors[i]
	})

	// try to get the Contents check the mirrors
	for _, base := range mirrors {
		url := fmt.Sprintf("%sContents-%s.gz", base, arch)
		// create a new request with the context
		// with this the cli avoids hanging
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			// something goes wrong, try with another mirror
			continue
		}
		// execute the request
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			// no error, and a OK, return the body
			return resp.Body, nil
		}
		// this point is reached if there is an error, or the code is not OK
		if resp != nil {
			// close the body to avoid leaking
			resp.Body.Close()
		}
	}
	// opps, none of the mirrors worked
	return nil, fmt.Errorf("failed to download Contents file for architecture '%s' from any available mirror", arch)
}

func parseContents(r io.Reader) (map[string]int, error) {
	// scan the contents
	scanner := bufio.NewScanner(r)
	// initialize the map to store the findings
	packages := make(map[string]int)

	for scanner.Scan() {
		// get the next line
		line := scanner.Text()
		// it is empty, ignore it
		if line == "" {
			continue
		}

		// get the last space
		// this is done like that because the file names have spaces
		// but the package don't, this way for sure the last space is the one that
		// marks the begining of the packages
		// I was doing some research to find which one was most performant
		// LastIndex or LastIndexByte
		// because I want the last space, this one is more efficient
		idx := strings.LastIndexByte(line, ' ')
		if idx == -1 {
			continue
		}
		// split by comma each package
		packageNames := strings.SplitSeq(line[idx+1:], ",")
		for pkg := range packageNames {
			pkg = strings.TrimSpace(pkg)
			// ignore empty package names
			if pkg != "" {
				packages[pkg]++
			}
		}
	}

	// something goes wrong, return the error
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// return the map with the counts
	return packages, nil
}

func calculateTopN(packages map[string]int, n int) []PackageCount {
	// be defensive, return empty if n has an invalid value
	if n <= 0 {
		return []PackageCount{}
	}

	// initialize the list based on the number of packages
	list := make([]PackageCount, 0, len(packages))
	// fill the list
	for name, count := range packages {
		list = append(list, PackageCount{Name: name, Count: count})
	}
	// sort the list
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})

	// return the top N packages
	// again be defensive, if N is bigger than the length of the list return the list
	limit := min(len(list), int(n))
	return list[:limit]
}

func printTopN(topN []PackageCount) {
	// show the result in a readable manner
	for i, pkg := range topN {
		fmt.Printf("%2d. %-50s %6d\n", i+1, pkg.Name, pkg.Count)
	}
}
