package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const policyName = "Cloud SaaS Incident Management"

// Strings to match on a page
const outOfCapacity = "is running low on spare capacity"
const ctorOutOfCapacity = "has logged a NotEnoughCapacity"
const allocatorsDown = "Website | Your site 'Allocators:"
const loggingDown = "Website | Your site 'Logging: production"
const metricsDown = "Website | Your site 'Metrics: production"
const indexFreshness = "Index freshness alert"
const terminatedOnHostError = "Instance(s) Terminated on Host Error"
const incidents = "Cloudbot's created a new incident"

type pageInfo struct {
	PageNumber  string
	Description string
	PolicyName  string
	CreatedOn   string
}

// extractPage maps csv line to a pageInfo
func extractPage(rawPageInfo []string) pageInfo {
	ii := pageInfo{
		PageNumber:  rawPageInfo[1],
		Description: rawPageInfo[2],
		PolicyName:  rawPageInfo[6],
		CreatedOn:   rawPageInfo[7],
	}
	return ii
}

func extractPageInfo(csvReader *csv.Reader) []pageInfo {
	var pageInfos []pageInfo
	for {
		line, error := csvReader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		page := extractPage(line)
		if page.PolicyName == policyName {
			pageInfos = append(pageInfos, page)
		}
	}
	return pageInfos
}

func getMatchPageCount(pageInfos []pageInfo, descriptionMatch string) int {
	count := 0
	for _, page := range pageInfos {
		if strings.Contains(page.Description, descriptionMatch) {
			count++
		}
	}
	return count
}

func printPageStats(pageInfos []pageInfo) {
	fmt.Printf("Total Alerts: %d\n", len(pageInfos))
	ooc := getMatchPageCount(pageInfos, outOfCapacity)
	ctorooc := getMatchPageCount(pageInfos, ctorOutOfCapacity)
	fmt.Printf("Total Capacity Alerts: %d\n", ooc+ctorooc)
	fmt.Printf("Out of Capacity Alerts: %d\n", ooc)
	fmt.Printf("Ctor Out of Capacity Alerts: %d\n", ctorooc)
	fmt.Printf("Bad Allocators: %d\n", getMatchPageCount(pageInfos, allocatorsDown))
	fmt.Printf("Total Incidents: %d\n", getMatchPageCount(pageInfos, incidents))
	fmt.Printf("Allocators on Old Templates: %d\n", getMatchPageCount(pageInfos, terminatedOnHostError))
	loggingMetricsDown := getMatchPageCount(pageInfos, loggingDown) + getMatchPageCount(pageInfos, metricsDown)
	fmt.Printf("Total Logging/Metrics: %d\n", loggingMetricsDown)
	fmt.Printf("Total Index Freshness: %d\n", getMatchPageCount(pageInfos, indexFreshness))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./pager-stats <pager-duty csv file>")
		os.Exit(1)
	}
	csvFileName := os.Args[1]
	csvFile, _ := os.Open(csvFileName)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	pageInfos := extractPageInfo(reader)
	printPageStats(pageInfos)
}
