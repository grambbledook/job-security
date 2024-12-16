package commands

import (
	"errors"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"pedeef/pedeef/asserts"
	"strings"
)

const (
	MergeUtilName = "merge"

	MergeOutputParamName      = "output"
	MergeOutputParamShortName = "o"
	MergeOutputParamUsage     = "Path to the output file"

	MergeInputParamName      = "input"
	MergeInputParamShortName = "i"
	MergeInputParamUsage     = "PDF files to merge"

	MergeDirectoryParamName      = "directory"
	MergeDirectoryParamShortName = "d"
	MergeDirectoryParamUsage     = "Directory with PDF files to merge"
)

var (
	NoInputParamsError       = errors.New("must provide at least one input file or directory")
	AmbigousInputError       = errors.New("ambiguous input, please provide either list of files of directory")
	EmptyInputDirectoryError = errors.New("input directory does not contain PDF files")
)

func NewMergeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: MergeUtilName,
		Run: mergePdf,
	}

	cmd.Flags().StringP(MergeOutputParamName, MergeOutputParamShortName, "", MergeOutputParamUsage)
	cmd.Flags().StringArrayP(MergeInputParamName, MergeInputParamShortName, nil, MergeInputParamUsage)
	cmd.Flags().StringP(MergeDirectoryParamName, MergeDirectoryParamShortName, "", MergeDirectoryParamUsage)

	return cmd
}

func mergePdf(cmd *cobra.Command, args []string) {

	out, err := cmd.Flags().GetString(MergeOutputParamName)
	asserts.NoError(err, "Output flag is required")

	in, err := cmd.Flags().GetStringArray(MergeInputParamName)
	asserts.NoError(err, "Input flag is required")

	dir, err := cmd.Flags().GetString(MergeDirectoryParamName)
	asserts.NoError(err, "Directory flag is required")

	files, err := getListOfFiles(in, dir)
	asserts.NoError(err, "Problem creating a list of files")

	fmt.Printf("Merging PDF files [%s]...\n", in)
	fmt.Printf("   Into [%s]...\n", out)

	tempFile, err := tempFile(out)
	asserts.NoError(err, "Error creating temporary file")
	defer os.Remove(tempFile.Name())

	config := model.NewDefaultConfiguration()
	config.Optimize = true

	err = api.MergeCreateFile(files, tempFile.Name(), false, config)
	asserts.NoError(err, "Merge failed")

	err = os.Rename(tempFile.Name(), out)
	asserts.NoError(err, "Failed moving tmp file to a final destination")

	fmt.Printf("Successfully merged files into %s\n", out)
}

func getListOfFiles(files []string, dir string) ([]string, error) {
	if dir == "" && len(files) == 0 {
		return nil, NoInputParamsError
	}

	if dir != "" && len(files) != 0 {
		return nil, AmbigousInputError
	}

	if dir == "" && len(files) != 0 {
		return files, nil
	}

	var candidates []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), "pdf") {
			candidates = append(candidates, path)
		}

		return nil
	})

	asserts.NoError(err, "Walk failed")

	if len(candidates) == 0 {
		return nil, EmptyInputDirectoryError
	}

	return candidates, nil
}

func tempFile(out string) (*os.File, error) {
	temp, err := os.CreateTemp("", fmt.Sprintf("%s.tmp", out))
	defer temp.Close()

	return temp, err
}
