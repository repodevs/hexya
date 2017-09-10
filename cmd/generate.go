// Copyright 2017 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"text/template"

	// We need to import models because of generated code
	_ "github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/loader"
)

const (
	// PoolDirRel is the name of the generated pool directory (relative to the hexya root)
	PoolDirRel string = "pool"
	// TempEmpty is the name of the temporary go file in the pool directory for startup
	TempEmpty string = "temp.go"
)

var generateCmd = &cobra.Command{
	Use:   "generate [projectDir]",
	Short: "Generate the source code of the model pool",
	Long: `Generate the source code of the pool package which includes the definition of all the models.
Additionally, this command creates the startup file of the project.
This command must be rerun after each source code modification, including module import.

  projectDir: the directory in which to find the go package that imports all the modules we want.
              If not set, projectDir defaults to the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := "."
		if len(args) > 0 {
			projectDir = args[0]
		}
		runGenerate(projectDir)
	},
}

var (
	generateEmptyPool bool
	testedModule      string
	importedPaths     []string
)

func init() {
	HexyaCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&testedModule, "test", "t", "", "Generate pool for testing the module in the given source directory. When set projectDir is ignored.")
	generateCmd.Flags().BoolVar(&generateEmptyPool, "empty", false, "Generate an empty pool package. When set projectDir is ignored.")
}

func runGenerate(projectDir string) {
	poolDir := filepath.Join(generate.HexyaDir, PoolDirRel)
	cleanPoolDir(poolDir)
	if generateEmptyPool {
		return
	}

	conf := loader.Config{
		AllowErrors: true,
	}

	fmt.Println(`Hexya Generate
------------`)
	fmt.Printf("Detected Hexya root directory at %s.\n", generate.HexyaDir)

	targetDir := filepath.Join(projectDir, "config")
	if testedModule != "" {
		targetDir, _ = filepath.Abs(testedModule)
	}
	fmt.Println("target dir", targetDir)
	importPack, err := build.ImportDir(targetDir, 0)
	if err != nil {
		panic(fmt.Errorf("Error while importing project: %s", err))
	}
	fmt.Printf("Project package found: %s.\n", importPack.Name)

	importedPaths = importPack.Imports
	if testedModule != "" {
		importedPaths = []string{importPack.ImportPath}
	}
	for _, ip := range importedPaths {
		conf.Import(ip)
	}

	fmt.Println(`Loading program...
Warnings may appear here, just ignore them if hexya-generate doesn't crash.`)

	program, _ := conf.Load()
	fmt.Println("Ok")

	fmt.Print("Generating pool...")
	generate.CreatePool(program, poolDir)
	fmt.Println("Ok")

	fmt.Print("Checking the generated code...")
	conf.AllowErrors = false
	_, err = conf.Load()
	if err != nil {
		fmt.Println("FAIL", err)
		os.Exit(1)
	}
	fmt.Println("Ok")

	fmt.Println("Pool generated successfully")
}

// cleanPoolDir removes all files in the given directory and leaves only
// one empty file declaring package 'pool'.
func cleanPoolDir(dirName string) {
	os.RemoveAll(dirName)
	os.MkdirAll(dirName, 0755)
	generate.CreateFileFromTemplate(filepath.Join(dirName, TempEmpty), emptyPoolTemplate, nil)
}

var emptyPoolTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package pool
`))
