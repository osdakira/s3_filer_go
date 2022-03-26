package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/apache/arrow/go/v8/parquet/file"
	// "github.com/apache/arrow/go/v8/parquet/metadata"
	// "github.com/apache/arrow/go/v8/parquet/schema"
)

func ReadParquet(buf []byte) (string, error) {
	r := bytes.NewReader(buf)
	rdr, err := file.NewParquetReader(r)
	// log.Println("err1", err)
	fileMetadata := rdr.MetaData()

	// log.Println("Version:", fileMetadata.Version())
	// log.Println("Created By:", fileMetadata.GetCreatedBy())
	// log.Println("Num Rows:", rdr.NumRows())

	// keyvaluemeta := fileMetadata.KeyValueMetadata()
	// if keyvaluemeta != nil {
	// 	fmt.Println("Key Value File Metadata:", keyvaluemeta.Len(), "entries")
	// 	keys := keyvaluemeta.Keys()
	// 	values := keyvaluemeta.Values()
	// 	for i := 0; i < keyvaluemeta.Len(); i++ {
	// 		log.Printf("Key nr %d %s: %s\n", i, keys[i], values[i])
	// 	}
	// }

	// log.Println("Number of RowGroups:", rdr.NumRowGroups())
	// log.Println("Number of Real Columns:", fileMetadata.Schema.Root().NumFields())
	// log.Println("Number of Columns:", fileMetadata.Schema.NumColumns())

	selectedColumns := []int{}

	if len(selectedColumns) == 0 {
		for i := 0; i < fileMetadata.Schema.NumColumns(); i++ {
			selectedColumns = append(selectedColumns, i)
		}
	} else {
		for _, c := range selectedColumns {
			if c < 0 || c >= fileMetadata.Schema.NumColumns() {
				fmt.Fprintln(os.Stderr, "selected column is out of range")
				os.Exit(1)
			}
		}
	}

	// fmt.Println("Number of Selected Columns:", len(selectedColumns))
	// for _, c := range selectedColumns {
	// 	descr := fileMetadata.Schema.Column(c)
	// 	log.Printf("Column %d: %s (%s", c, descr.Path(), descr.PhysicalType())
	// 	if descr.ConvertedType() != schema.ConvertedTypes.None {
	// 		log.Printf("/%s", descr.ConvertedType())
	// 		if descr.ConvertedType() == schema.ConvertedTypes.Decimal {
	// 			dec := descr.LogicalType().(*schema.DecimalLogicalType)
	// 			log.Printf("(%d,%d)", dec.Precision(), dec.Scale())
	// 		}
	// 	}
	// 	log.Print(")\n")
	// }

	var str_build strings.Builder

	for r := 0; r < rdr.NumRowGroups(); r++ {
		// log.Println("--- Row Group:", r, " ---")

		rgr := rdr.RowGroup(r)
		// rowGroupMeta := rgr.MetaData()
		// log.Println("--- Total Bytes:", rowGroupMeta.TotalByteSize(), " ---")
		// log.Println("--- Rows:", rgr.NumRows(), " ---")

		// for _, c := range selectedColumns {
		// chunkMeta, err := rowGroupMeta.ColumnChunk(c)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Println("Column", c)
		// if set, _ := chunkMeta.StatsSet(); set {
		// 	stats, err := chunkMeta.Statistics()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	// fmt.Printf(" Values: %d", chunkMeta.NumValues())
		// 	if stats.HasMinMax() {
		// 		// log.Printf(", Min: %v, Max: %v",
		// 		// 	metadata.GetStatValue(stats.Type(), stats.EncodeMin()),
		// 		// 	metadata.GetStatValue(stats.Type(), stats.EncodeMax()))
		// 	}
		// 	if stats.HasNullCount() {
		// 		// log.Printf(", Null Values: %d", stats.NullCount())
		// 	}
		// 	if stats.HasDistinctCount() {
		// 		// log.Printf(", Distinct Values: %d", stats.DistinctCount())
		// 	}
		// 	// log.Println()
		// } else {
		// 	// log.Println(" Values:", chunkMeta.NumValues(), "Statistics Not Set")
		// }

		// log.Print(" Compression: ", chunkMeta.Compression())
		// log.Print(", Encodings:")
		// for _, enc := range chunkMeta.Encodings() {
		// 	log.Print(" ", enc)
		// }
		// log.Println()

		// log.Print(" Uncompressed Size: ", chunkMeta.TotalUncompressedSize())
		// log.Println(", Compressed Size:", chunkMeta.TotalCompressedSize())
		// }

		// if config.OnlyMetadata {
		// 	continue
		// }

		// fmt.Println("--- Values ---")

		const colwidth = 18

		scanners := make([]*Dumper, len(selectedColumns))
		for idx, c := range selectedColumns {
			scanners[idx] = createDumper(rgr.Column(c))
			str_build.WriteString(fmt.Sprintf(fmt.Sprintf("%%-%ds|", colwidth), rgr.Column(c).Descriptor().Name()))
			// log.Printf(fmt.Sprintf("%%-%ds|", colwidth), rgr.Column(c).Descriptor().Name())
		}
		str_build.WriteString("\n")

		for i := 0; i < 10; i++ {
			data := false
			for _, s := range scanners {
				if val, ok := s.Next(); ok {
					str_build.WriteString(s.FormatValue(val, colwidth) + "|")
					// log.Print(s.FormatValue(val, colwidth), "|")
					data = true
				} else {
					str_build.WriteString(fmt.Sprintf(fmt.Sprintf("%%-%ds|", colwidth), ""))
					// log.Printf(fmt.Sprintf("%%-%ds|", colwidth), "")
				}
			}
			str_build.WriteString("\n")
			// log.Println()
			if !data {
				break
			}
		}
		str_build.WriteString("\n")
	}

	return str_build.String(), err
}
