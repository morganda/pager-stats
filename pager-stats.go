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
const oldctorOutOfCapacity = "has logged a NotEnoughCapacity"
const ctorOutOfCapacity = "cannot find enough available allocator capacity"
const allocatorsDown = "Website | Your site 'Allocators:"
const soteriaallocatorsDown = "Soteria :: Allocator is Unhealthy"
const esspallocatorsDown = "for check 'Allocators'"
const loggingDown = "Website | Your site 'Logging:"
const metricsDown = "Website | Your site 'Metrics:"
const monitorDown = "Website | Your site 'Monitor:"
const othersDown = "went down"
const esspothersDown = "Heartbeat Alert"
const indexFreshness = "Index freshness alert"
const terminatedOnHostError = "Instance(s) Terminated on Host Error"
const incidents = "Cloudbot's created a new incident"
const zookeeperDisk = "Sent bytes for cloud-production-168820 director"

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
	oldctorooc := getMatchPageCount(pageInfos, oldctorOutOfCapacity)
	ctorooc := getMatchPageCount(pageInfos, ctorOutOfCapacity)
	othersdown := getMatchPageCount(pageInfos, othersDown)
	esspothersdown := getMatchPageCount(pageInfos, esspothersDown)
	allocatorsdown := getMatchPageCount(pageInfos, allocatorsDown)
	soteriaallocatorsdown := getMatchPageCount(pageInfos, soteriaallocatorsDown)
	esspallocatorsdown := getMatchPageCount(pageInfos, esspallocatorsDown)
	alldown := othersdown + esspothersdown
	allallocators := esspallocatorsdown + allocatorsdown + soteriaallocatorsdown
	loggingMetricsDown := getMatchPageCount(pageInfos, loggingDown) + getMatchPageCount(pageInfos, metricsDown) + getMatchPageCount(pageInfos, monitorDown)

	template := "%d pages in the past month, which break down to %d related to capacity (0 needing more capacity, %d constructor out of capacity), %d pages for failed allocators, %d GCP allocators rebuilt with old templates, %d pages for incidents, %d for logging or metrics down, %d for index freshness and %d pages for non-allocator host failures.\n"
	fmt.Printf(template, len(pageInfos), ctorooc+oldctorooc, ctorooc+oldctorooc, allallocators, getMatchPageCount(pageInfos, terminatedOnHostError), getMatchPageCount(pageInfos, incidents), loggingMetricsDown, getMatchPageCount(pageInfos, indexFreshness), alldown-loggingMetricsDown-allallocators)
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
