package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type config struct {
	startVal  int
	count     int
	increment int
	unit      int
	unitName  string
	path      string
	tenant    string
	filter    string
	subject   string
	attr      string
}

func main() {

	var cfg config

	flag.IntVar(&cfg.startVal, "start_value", 0, "from what value to start generating")
	flag.IntVar(&cfg.count, "count", 3, "number of RatingProfiles to generate")
	flag.IntVar(&cfg.increment, "increment", 102400, "increment value")
	flag.IntVar(&cfg.unit, "unit", 1024, "unit value for conversion")
	flag.StringVar(&cfg.unitName, "unit_name", "MB", "unit name")
	flag.StringVar(&cfg.path, "path", "", "where to put the generated tps")
	flag.StringVar(&cfg.tenant, "tenant", "cgrates.org", "tenant")
	flag.StringVar(&cfg.filter, "filter", "*string:~*req.Account:1001", "attribute profile filter")
	flag.StringVar(&cfg.subject, "subject", "main_balance_subj", "subject")
	flag.StringVar(&cfg.attr, "attr", "ap_rating", "attr")
	flag.Parse()

	if cfg.increment%cfg.unit != 0 || cfg.startVal%cfg.unit != 0 {
		log.Fatal("increment and start_value must be multiples of unit_value")
	}

	if cfg.path != "" {
		if err := os.MkdirAll(cfg.path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	writeFile := func(filename, content string) {
		if cfg.path == "" {
			fmt.Println(content)
		} else {
			filePath := filepath.Join(cfg.path, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				log.Fatal(err)
			}
		}
	}

	val := 0
	ratingProfilesContent := "#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject"
	for range cfg.count {
		convVal := fmt.Sprintf("%d%s", val/cfg.unit, cfg.unitName)
		ratingProfilesContent += fmt.Sprintf("\n%s,GTE_%s,%s,,RP_GTE_%s,", cfg.tenant, convVal, cfg.subject, convVal)
		val += cfg.increment
	}
	writeFile("RatingProfiles.csv", ratingProfilesContent)

	val = 0
	ratingPlansContent := "#Id,DestinationRatesId,TimingTag,Weight"
	for range cfg.count {
		convVal := fmt.Sprintf("%d%s", val/cfg.unit, cfg.unitName)
		ratingPlansContent += fmt.Sprintf("\nRP_GTE_%s,DR_GTE_%s,*any,", convVal, convVal)
		val += cfg.increment
	}
	writeFile("RatingPlans.csv", ratingPlansContent)

	val = 0
	ratesContent := "#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart"
	ratesContent += fmt.Sprintf("\nRT_GTE_%s,0,0,1,1,0", fmt.Sprintf("0%s", cfg.unitName))
	for range cfg.count - 1 {
		val += cfg.increment
		convVal := fmt.Sprintf("%d%s", val/cfg.unit, cfg.unitName)
		ratesContent += fmt.Sprintf("\nRT_GTE_%s,0,0,%d,%d,0\n", convVal, val, val)
		ratesContent += fmt.Sprintf("RT_GTE_%s,0,0,1,1,%d", convVal, val)
	}
	writeFile("Rates.csv", ratesContent)

	val = 0
	destinationRatesContent := "#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy"
	for range cfg.count {
		convVal := fmt.Sprintf("%d%s", val/cfg.unit, cfg.unitName)
		destinationRatesContent += fmt.Sprintf("\nDR_GTE_%s,*any,RT_GTE_%s,*up,,,", convVal, convVal)
		val += cfg.increment
	}
	writeFile("DestinationRates.csv", destinationRatesContent)

	val = 0
	attributesContent := "#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight"
	attributesContent += fmt.Sprintf("\n%s,%s,,%s,,,,,,,", cfg.tenant, cfg.attr, cfg.filter)
	for i := range cfg.count {
		convVal := fmt.Sprintf("%d%s", val/cfg.unit, cfg.unitName)
		switch i {
		case 0:
			attributesContent += fmt.Sprintf("\n%s,%s,,,,*lt:~*req.Usage:%d,*req.Category,*constant,GTE_%s,,", cfg.tenant, cfg.attr, val+cfg.increment, convVal)
		case cfg.count - 1:
			attributesContent += fmt.Sprintf("\n%s,%s,,,,*gte:~*req.Usage:%d,*req.Category,*constant,GTE_%s,,", cfg.tenant, cfg.attr, val, convVal)
		default:
			attributesContent += fmt.Sprintf("\n%s,%s,,,,*gte:~*req.Usage:%d;*lt:~*req.Usage:%d,*req.Category,*constant,GTE_%s,,", cfg.tenant, cfg.attr, val, val+cfg.increment, convVal)
		}
		val += cfg.increment
	}
	writeFile("Attributes.csv", attributesContent)
}
