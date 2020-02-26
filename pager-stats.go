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
const outOfCapacityGCP = "Instance family gcp."
const outOfCapacityAzure = "Instance family azure."
const outOfCapacityIBM = "Instance family ibm."
const ctorOutOfCapacity = "has logged a NotEnoughCapacity"
const allocatorsDown = "Soteria detected that the allocator"
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
	fmt.Printf("Total Alerts: %d\n", len(pageInfos))
	ooc := getMatchPageCount(pageInfos, outOfCapacity)
	oocgcp := getMatchPageCount(pageInfos, outOfCapacityGCP)
	oocazure := getMatchPageCount(pageInfos, outOfCapacityAzure)
	oocibm := getMatchPageCount(pageInfos, outOfCapacityIBM)
	ctorooc := getMatchPageCount(pageInfos, ctorOutOfCapacity)
	othersdown := getMatchPageCount(pageInfos, othersDown)
	esspothersdown := getMatchPageCount(pageInfos, esspothersDown)
	allocatorsdown := getMatchPageCount(pageInfos, allocatorsDown)
	esspallocatorsdown := getMatchPageCount(pageInfos, esspallocatorsDown)
	alldown := othersdown + esspothersdown
	allallocators := esspallocatorsdown + allocatorsdown
	fmt.Printf("Total Capacity Alerts: %d\n", ooc+ctorooc)
	fmt.Printf("Out of Capacity Alerts: %d\n", ooc)
	fmt.Printf("Out of Capacity Alerts GCP: %d\n", oocgcp)
	fmt.Printf("Out of Capacity Alerts Azure: %d\n", oocazure)
	fmt.Printf("Out of Capacity Alerts IBM: %d\n", oocibm)
	fmt.Printf("Out of Capacity Alerts AWS: %d\n", ooc - oocgcp - oocazure - oocibm)
	fmt.Printf("Ctor Out of Capacity Alerts: %d\n", ctorooc)
	fmt.Printf("Total Zookeeper Disk Alerts: %d\n", getMatchPageCount(pageInfos, zookeeperDisk))
	fmt.Printf("Bad Allocators: %d\n", allocatorsdown + esspallocatorsdown)
	fmt.Printf("Allocators on Old Templates: %d\n", getMatchPageCount(pageInfos, terminatedOnHostError))
	fmt.Printf("Total Incidents: %d\n", getMatchPageCount(pageInfos, incidents))
	loggingMetricsDown := getMatchPageCount(pageInfos, loggingDown) + getMatchPageCount(pageInfos, metricsDown) + getMatchPageCount(pageInfos, monitorDown)
	fmt.Printf("Total Logging/Metrics: %d\n", loggingMetricsDown)
	fmt.Printf("Total Index Freshness: %d\n", getMatchPageCount(pageInfos, indexFreshness))
	fmt.Printf("Non-allocator-failures: %d\n", alldown-loggingMetricsDown-allallocators)
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
