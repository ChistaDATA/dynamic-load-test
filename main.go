package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Task struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

func readTaskFiles(filePaths []string) ([]Task, error) {
	var tasks []Task
	for _, filePath := range filePaths {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var fileTasks []Task
		err = yaml.Unmarshal(data, &fileTasks)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, fileTasks...)
	}
	return tasks, nil
}

func extractFields(tasks []Task) (map[string][]Task, map[string]int, map[string]int) {
	nonAggregatedFieldsMap := make(map[string][]Task)
	aggregatedFields := make(map[string]int)
	whereFieldsSet := make(map[string]int)

	selectPattern := regexp.MustCompile(`(?i)SELECT (.+?) FROM`)
	aggregatePattern := regexp.MustCompile(`(?i)(SUM|COUNT|AVG|MIN|MAX)\((.+?)\)`)
	wherePattern := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:GROUP BY|ORDER BY|LIMIT|$)`)

	for _, task := range tasks {
		query := task.Query
		selectMatches := selectPattern.FindStringSubmatch(query)
		aggregateMatches := aggregatePattern.FindAllStringSubmatch(query, -1)
		whereMatches := wherePattern.FindStringSubmatch(query)

		if len(selectMatches) > 1 {
			fields := strings.Split(selectMatches[1], ",")
			var nonAggregatedFields []string
			for _, field := range fields {
				field = strings.TrimSpace(field)
				if field == "now()" {
					continue
				}
				if !aggregatePattern.MatchString(field) {
					nonAggregatedFields = append(nonAggregatedFields, field)
				} else {
					match := aggregatePattern.FindStringSubmatch(field)
					if len(match) > 2 {
						aggregatedFields[match[2]] = aggregatedFields[match[2]] + 1
					}
				}
			}
			key := strings.Join(nonAggregatedFields, "__")
			nonAggregatedFieldsMap[key] = append(nonAggregatedFieldsMap[key], task)
		}

		for _, match := range aggregateMatches {
			field := strings.TrimSpace(match[2])
			aggregatedFields[field] = aggregatedFields[field] + 1
		}

		if len(whereMatches) > 1 {
			whereFields := extractFieldsFromConditions(whereMatches[1])
			for _, field := range whereFields {
				whereFieldsSet[field] = whereFieldsSet[field] + 1
			}
		}
	}
	return nonAggregatedFieldsMap, aggregatedFields, whereFieldsSet
}

func extractFieldsFromConditions(conditions string) []string {
	conditionPattern := regexp.MustCompile(`\b(\w+)\b\s*(?:=|!=|<|<=|>|>=|IN|BETWEEN|LIKE)`)
	matches := conditionPattern.FindAllStringSubmatch(conditions, -1)

	var fields []string
	for _, match := range matches {
		fields = append(fields, match[1])
	}

	return fields
}

func writeQueriesToFiles(nonAggregatedFieldsMap map[string][]Task, daysTable []int) error {
	for key, tasks := range nonAggregatedFieldsMap {
		for _, task := range tasks {
			days := extractDaysFromWhereClause(task.Query)
			daysForFileName := getNumberOfDayForFile(days, daysTable)
			fileName := fmt.Sprintf("task_%s_%d_days.yaml", key, daysForFileName)
			task.Name = fmt.Sprintf("%s_%d_days", key, daysForFileName)
			data, err := yaml.Marshal([]Task{task})
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = appendLineToFile(fileName, string(data))
			if err != nil {
				fmt.Printf("Error appending line to file: %v\n", err)
				return err
			}
		}
	}
	return nil
}

func appendLineToFile(fileName, line string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(line); err != nil {
		return err
	}
	return nil
}

func getNumberOfDayForFile(days int, daysTable []int) int {
	for _, daysInFile := range daysTable {
		if days <= daysInFile {
			return daysInFile
		}
	}
	return 99
}

func extractDaysFromWhereClause(query string) int {
	wherePattern := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:GROUP BY|ORDER BY|LIMIT|$)`)
	matches := wherePattern.FindStringSubmatch(query)
	if len(matches) > 1 {
		conditions := matches[1]
		datePattern := regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2})\b`)
		dateMatches := datePattern.FindAllString(conditions, -1)
		if len(dateMatches) == 2 {
			startDate, err1 := time.Parse("2006-01-02", dateMatches[0])
			endDate, err2 := time.Parse("2006-01-02", dateMatches[1])
			if err1 != nil || err2 != nil {
				log.Printf("Error parsing dates: %v, %v", err1, err2)
				return 777
			}
			days := int(endDate.Sub(startDate).Hours() / 24)
			return days
		}
	}
	return 0
}

func main() {
	filePaths := []string{"task01.yaml", "task02.yaml", "task03.yaml", "task04.yaml", "task05.yaml", "task06.yaml"}
	tasks, err := readTaskFiles(filePaths)
	if err != nil {
		log.Fatalf("Error reading task files: %v", err)
	}

	nonAggregatedFieldsMap, aggregatedFields, whereFields := extractFields(tasks)
	keys := make([]string, 0, len(nonAggregatedFieldsMap))
	for key := range nonAggregatedFieldsMap {
		keys = append(keys, key)
	}
	fmt.Printf("Non-aggregated fields: %v\n", strings.Join(keys, ", "))
	fmt.Printf("Aggregated fields: %v\n", aggregatedFields)
	fmt.Printf("WHERE clause fields: %v\n", whereFields)

	daysTable := []int{
		3,
		10,
		32,
		124,
		185,
	}

	err = writeQueriesToFiles(nonAggregatedFieldsMap, daysTable)
	if err != nil {
		log.Fatalf("Error writing queries to files: %v", err)
	}
}
