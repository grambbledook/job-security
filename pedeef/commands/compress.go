package commands

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"iter"
	"os"
	"path/filepath"
	"pedeef/asserts"
	"strings"
)

const (
	CompressUtilName = "compress"

	CompressOutputParamName      = "output"
	CompressOutputParamShortName = "o"
	CompressOutputParamUsage     = "Path to the output file"

	CompressInputParamName      = "input"
	CompressInputParamShortName = "i"
	CompressInputParamUsage     = "PDF file to compress"

	CompressionParamName   = "compression"
	CompressParamShortName = "c"
	CompressParamUsage     = "Compression percentage"

	QualityParamName      = "quality"
	QualityParamShortName = "q"
	QualityParamUsage     = "Quality percentage"

	CompressNoTempFileParamName      = "no-temp-file"
	CompressNoTempFileParamShortName = "n"
	CompressNoTempFileParamUsage     = "Don't write to temp file"
)

func NewCompressCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: CompressUtilName,
		Run: compressPdf,
	}

	cmd.Flags().StringP(CompressOutputParamName, CompressOutputParamShortName, "", CompressOutputParamUsage)
	cmd.Flags().StringP(CompressInputParamName, CompressInputParamShortName, "", CompressInputParamUsage)
	cmd.Flags().IntP(CompressionParamName, CompressParamShortName, 100, CompressParamUsage)
	cmd.Flags().IntP(QualityParamName, QualityParamShortName, 75, QualityParamUsage)
	cmd.Flags().BoolP(CompressNoTempFileParamName, CompressNoTempFileParamShortName, false, CompressNoTempFileParamUsage)

	return cmd
}

type TempDir struct {
	name string
}

func NewTempDir() *TempDir {
	dir := &TempDir{
		name: uuid.NewString(),
	}
	err := os.MkdirAll(dir.name, os.ModePerm)
	asserts.NoError(err, fmt.Sprintf("Error creating temp dir: %s", dir.name))
	return dir
}

func (t *TempDir) Clean() {
	err := os.RemoveAll(t.name)
	asserts.NoError(err, fmt.Sprintf("Error cleaning temp dir: %s", t.name))
}

func (t *TempDir) jpegs() iter.Seq2[string, string] {

	return func(yield func(string, string) bool) {
		_ = filepath.Walk(t.name, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if strings.HasSuffix(info.Name(), "jpeg") {
				if !yield(path, info.Name()) {
					return nil
				}
			}

			if strings.HasSuffix(info.Name(), "jpg") {
				if !yield(path, info.Name()) {
					return nil
				}
			}

			return nil
		})
	}
}

func compressPdf(cmd *cobra.Command, args []string) {

	out, _ := cmd.Flags().GetString(CompressOutputParamName)
	asserts.NotEmpty(out, "Output flag is required")

	in, _ := cmd.Flags().GetString(CompressInputParamName)
	asserts.NotEmpty(in, "Input flag is required")

	compression, _ := cmd.Flags().GetInt(CompressionParamName)
	asserts.NotZeru(compression, "Compression flag is required")

	quality, _ := cmd.Flags().GetInt(QualityParamName)
	asserts.NotZeru(compression, "Quality flag is required")

	noTemp, err := cmd.Flags().GetBool(NoTempFileParamName)
	asserts.NoError(err, "NoTempFile flag is required")

	fmt.Printf("Compressing PDF file [%s]...\n", in)
	fmt.Printf("   Into [%s]...\n", out)
	fmt.Printf("   Compression [%d]...\n", compression)
	fmt.Printf("   JPEG Quality [%d]...\n", quality)

	file := out
	if !noTemp {
		tempFile, err := tempFile(out)
		asserts.NoError(err, "Error creating temporary file")
		defer os.Remove(tempFile.Name())

		file = tempFile.Name()
	}
	process(in, file, compression, quality)

	err = os.Rename(file, out)
	asserts.NoError(err, "Failed moving tmp file to a final destination")

	fmt.Printf("Successfully compressd files into %s\n", out)
}

func process(in, out string, compression, quality int) {
	originals := NewTempDir()
	defer originals.Clean()

	compressed := NewTempDir()
	defer compressed.Clean()

	extractImages(in, originals)

	for src, name := range originals.jpegs() {
		dest := filepath.Join(compressed.name, name)
		compressImage(src, dest, compression, quality)
	}

	compileDocument(compressed, out)
}

func extractImages(in string, outputDir *TempDir) {
	config := model.NewDefaultConfiguration()
	err := api.ExtractImagesFile(in, outputDir.name, nil, config)
	asserts.NoError(err, "Error extracting images")

	fmt.Printf("Images extracted successfully to: %s\n", outputDir.name)
}

func compressImage(src, dst string, compression, quality int) {
	fmt.Printf("Compressing file [%s]...\n", src)
	fmt.Printf("  Destination file [%s]...\n", dst)

	file, err := os.Open(src)
	asserts.NoError(err, "Error opening source file")
	defer file.Close()

	img, _, err := image.Decode(file)
	asserts.NoError(err, "Error decoding image")

	out, err := os.Create(dst)
	asserts.NoError(err, "Error creating destination file")

	scale := float64(compression) / float64(100)

	targetWidth := float64(img.Bounds().Dx()) * scale
	targetHeight := float64(img.Bounds().Dy()) * scale
	fmt.Printf("  Target resolution: %0.f x %0.f\n", targetHeight, targetWidth)

	resizedImg := resize.Resize(uint(targetWidth), uint(targetHeight), img, resize.Lanczos3)

	err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: quality})
	asserts.NoError(err, "Error encoding image")
	defer out.Close()

	fmt.Printf(" Image resized and compressed successfully! Saved to: %s\n", dst)
}

func compileDocument(src *TempDir, dst string) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for path, _ := range src.jpegs() {
		pdf.AddPage()

		scaleWidth := gopdf.PageSizeA4.W
		scaleHeight := gopdf.PageSizeA4.H

		fmt.Printf("  Scale factor for page: %s\n", path)
		fmt.Printf("  %f x %f\n", scaleWidth, scaleHeight)

		err := pdf.Image(path, 0, 0, &gopdf.Rect{W: scaleWidth, H: scaleHeight})
		asserts.NoError(err, "Error creating image")
	}

	err := pdf.WritePdf(dst)
	asserts.NoError(err, "Error creating pdf file")

	fmt.Println("PDF created successfully")
}
