## Build
```
go build -o pedeef
```

## Compress image-only PDF file 
```
Usage:
  pedeef compress [flags]

Flags:
  -c, --compression int   Compression percentage (default 100)
  -i, --input string      PDF file to compress
  -o, --output string     Path to the output file
  -q, --quality int       Quality percentage (default 75)
```

```
pedeef compress -i huge.pdf -o compressed.pdf -c 65 -q 65
```

## Merge multiple PDF files into one
```
Usage:
  pedeef merge [flags]

Flags:
  -d, --directory string    Directory with PDF files to merge
  -i, --input stringArray   PDF files to merge
  -o, --output string       Path to the output file
```

```
pedeef merge -o combined.pdf -d all_the_pdfs
```

