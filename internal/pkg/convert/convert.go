package convert

import (
	"errors"
	"fmt"
	"os"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/fetch"

	"github.com/bomctl/bomctl/internal/pkg/utils"
)

var errUnsupportedURL = errors.New("unsupported URL scheme")

func Exec(sbomURL, outputFile, outputBomFormat, outputEncodingFormat string, useNetRC bool) error {
	logger := utils.NewLogger("convert")
	if outputFile == "" {
		return fmt.Errorf("output file not found")
	}

	document, detectedFormat, err := fetch.Document(sbomURL, outputFile, useNetRC, logger)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	err = db.AddDocument(document)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	targetFormat, err := utils.GetConvertSBOMFormat(outputBomFormat, outputEncodingFormat, detectedFormat)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	logger.Info("Converting SBOM", "From", detectedFormat, "To", targetFormat)

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer out.Close()

	utils.WriteSBOM(document, targetFormat, out)

	logger.Info("SBOM Written", "File", outputFile)

	return nil
}
