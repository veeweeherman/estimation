package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type config struct {
	Github struct {
		Password string `yaml:"password"`
		Owner    string `yaml:"owner"`
		Repo     string `yaml:"repo"`
	} `yaml:"github"`
	Columns []struct {
		Name string `yaml:"name"`
	} `yaml:"columns"`
	SizeLabels []struct {
		Name string `yaml:"name"`
	} `yaml:"size_labels"`
	EngineeringFunctions []struct {
		Name string `yaml:"name"`
	} `yaml:"engineering_functions"`
}

func main() {
	// read from config.yml
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	// declare config object to read values from it
	var cfg config

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println(err)
	}

	// Github Go client setup
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.Password},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// List projects for given user and repo
	projects, _, err := client.Repositories.ListProjects(ctx, cfg.Github.Owner, cfg.Github.Repo, nil)

	// Target project id
	projectID := projects[0].GetID()

	// All columns from target project
	columns, _, err := client.Projects.ListProjectColumns(ctx, projectID, nil)
	// fmt.Println("All Project Columns: ", columns)

	// Target column names
	// fmt.Println("column names from config: ", cfg.Columns)

	targetColumns := filterTargetColumns(cfg, columns)
	// fmt.Println("targetColumns: ", targetColumns)

	// All cards from target columns
	cardsForTargetColumns := make([]*github.ProjectCard, 0)

	for i := 0; i < len(targetColumns); i++ {
		cards, _, err := client.Projects.ListProjectCards(ctx, targetColumns[i].GetID(), nil)
		if err != nil {
			fmt.Println(err)
		}
		for j := 0; j < len(cards); j++ {
			cardsForTargetColumns = append(cardsForTargetColumns, cards[j])
		}
	}

	// Fetching all size labels
	sizeLabels := make([]*github.Label, 0, len(cfg.SizeLabels))
	for i := 0; i < len(cfg.SizeLabels); i++ {
		label, _, err := client.Issues.GetLabel(ctx, cfg.Github.Owner, cfg.Github.Repo, cfg.SizeLabels[i].Name)
		if err != nil {
			fmt.Println(err)
		}
		sizeLabels = append(sizeLabels, label)
	}

	// Fetching all engineering functions from config
	engineeringFunctions := make([]string, 0, len(cfg.EngineeringFunctions))
	for i := 0; i < len(cfg.EngineeringFunctions); i++ {
		engFunction := cfg.EngineeringFunctions[i].Name
		engineeringFunctions = append(engineeringFunctions, engFunction)
	}

	// Creating Label frequency map
	labelFrequency := map[string]map[string]int{}
	for i := 0; i < len(engineeringFunctions); i++ {
		sizes := map[string]int{}
		for j := 0; j < len(cfg.SizeLabels); j++ {
			sizes[cfg.SizeLabels[j].Name] = 0
		}
		labelFrequency[engineeringFunctions[i]] = sizes
	}
	fmt.Println(labelFrequency)
}

func filterTargetColumns(conf config, columns []*github.ProjectColumn) []*github.ProjectColumn {
	// fmt.Println("Number of target columns from config: ", len(conf.Columns))

	result := make([]*github.ProjectColumn, 0, len(conf.Columns))

	for i := 0; i < len(columns); i++ {
		for j := 0; j < len(conf.Columns); j++ {
			if columns[i].GetName() == conf.Columns[j].Name {
				// fmt.Println(columns[i].GetName(), " == ", conf.Columns[j].Name)
				result = append(result, columns[i])
			}
		}
	}

	return result
}
