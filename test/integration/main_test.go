package integration

import (
	"flag"
	"fmt"
	"strings"

	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/kr/pretty"
	color "github.com/logrusorgru/aurora"
	"github.com/otiai10/copy"
)

const binaryName = "mani"

var tmpPath = "/tmp"
var rootDir = ""
var goldenDir = filepath.Join("./test", "integration", "golden")

var debug = flag.Bool("debug", false, "debug")
var update = flag.Bool("update", false, "update golden files")
var clean = flag.Bool("clean", false, "Clean tmp directory after run")

var copyOpts = copy.Options{
	Skip: func(src string) (bool, error) {
		return strings.HasSuffix(src, ".git"), nil
	},
}

type TemplateTest struct {
	TestName   string
	InputFiles []string
	TestCmd    string
	Golden     string
	WantErr    bool
}

type TestFile struct {
	t    *testing.T
	name string
	dir  string
}

func NewGoldenFile(t *testing.T, name string) *TestFile {
	return &TestFile{t: t, name: "stdout.golden", dir: filepath.Join("golden", name)}
}

func (tf *TestFile) Dir() string {
	tf.t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tf.t.Fatal("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), tf.dir)
}

func (tf *TestFile) path() string {
	tf.t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tf.t.Fatal("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), tf.dir, tf.name)
}

func (tf *TestFile) Write(content string) {
	tf.t.Helper()
	err := os.MkdirAll(filepath.Dir(tf.path()), os.ModePerm)
	if err != nil {
		tf.t.Fatalf("could not create directory %s: %v", tf.name, err)
	}

	err = ioutil.WriteFile(tf.path(), []byte(content), 0644)
	if err != nil {
		tf.t.Fatalf("could not write %s: %v", tf.name, err)
	}
}

func (tf *TestFile) AsFile() *os.File {
	tf.t.Helper()
	file, err := os.Open(tf.path())
	if err != nil {
		tf.t.Fatalf("could not open %s: %v", tf.name, err)
	}
	return file
}

func (tf *TestFile) load() string {
	tf.t.Helper()

	content, err := ioutil.ReadFile(tf.path())
	if err != nil {
		tf.t.Fatalf("could not read file %s: %v", tf.name, err)
	}

	return string(content)
}

func clearGolden(goldenDir string) {
	// Guard against accidently deleting outside directory
	if strings.Contains(goldenDir, "golden") {
		os.RemoveAll(goldenDir)
	}
}

func clearTmp() {
	dir, _ := ioutil.ReadDir(path.Join(tmpPath, "golden"))
	for _, d := range dir {
		os.RemoveAll(path.Join(tmpPath, "golden", path.Join([]string{d.Name()}...)))
	}
}

func diff(expected, actual interface{}) []string {
	return pretty.Diff(expected, actual)
}

// 1. Clean tmp directory
// 2. Create mani binary
// 3. cd into test/tmp
func TestMain(m *testing.M) {
	fmt.Println("----------------------")
	fmt.Println("LALALA")
	fmt.Println("----------------------")
	clearTmp()
}

func countFilesAndFolders(dir string) int {
	var count = 0
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}

			count = count + 1

			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		fmt.Printf("could not walk dir: %v", err)
		os.Exit(1)
	}

	return count
}

func Run(t *testing.T, tt TemplateTest) {
	var tmpDir = filepath.Join(tmpPath, "golden", tt.Golden)
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err = os.MkdirAll(tmpDir, os.ModePerm)
		if err != nil {
			fmt.Printf("could not create directory at %s: %v", tmpPath, err)
			os.Exit(1)
		}
	}

	err := os.Chdir(tmpDir)
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	var fixturesDir = filepath.Join(rootDir, "fixtures")

	t.Cleanup(func() {
		if *clean == true {
			clearTmp()
		}
	})

	// Copy fixture files
	for _, file := range tt.InputFiles {
		var configPath = filepath.Join(fixturesDir, file)
		err := copy.Copy(configPath, filepath.Base(file), copyOpts)

		if err != nil {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
			os.Exit(1)
		}
	}

	// Run test command
	cmd := exec.Command("sh", "-c", tt.TestCmd)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()

	// TEST: Check we get error if we want error
	if (err != nil) != tt.WantErr {
		t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.WantErr, err != nil, err)
	}

	if *debug {
		fmt.Println(tt.TestCmd)
		fmt.Println(string(output))
	}

	// Save output from command as golden file
	golden := NewGoldenFile(t, tt.Golden)
	actual := string(output)

	var goldenFile = path.Join(tmpDir, "stdout.golden")
	err = ioutil.WriteFile(goldenFile, []byte(actual), 0644)
	if err != nil {
		t.Fatalf("could not write %s: %v", goldenFile, err)
	}

	if *update {
		clearGolden(golden.Dir())

		// Write stdout of test command to golden file
		golden.Write(actual)

		err := copy.Copy(tmpDir, golden.Dir(), copyOpts)
		if err != nil {
			fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
			os.Exit(1)
		}
	} else {
		err := filepath.Walk(golden.Dir(), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if path == tmpDir {
				return nil
			}

			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			tmpPath := filepath.Join(tmpDir, filepath.Base(path))

			actual, err := ioutil.ReadFile(tmpPath)
			expected, err := ioutil.ReadFile(path)

			// TEST: Check file content difference for each generated file
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("\nfile: %v\ndiff: %v", color.Blue(path), color.Red(diff(expected, actual)))
			}

			return nil
		})

		// TEST: Check the total amount of files and directories match
		expectedCount := countFilesAndFolders(golden.Dir())
		actualCount := countFilesAndFolders(tmpDir)

		if expectedCount != actualCount {
			t.Fatalf("\nexpected count: %v\nactual count: %v", color.Green(expectedCount), color.Red(actualCount))
		}

		if err != nil {
			t.Fatalf("Error: %v", color.Red(err))
		}
	}
}
