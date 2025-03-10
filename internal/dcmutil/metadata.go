package dcmutil

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"

	"bytes"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// Helper method to extract DICOM metadata
func ExtractDICOMData(filepath string) (dicom.Dataset, error) {
	dicom_dataset, err := dicom.ParseFile(filepath, nil)
	if err != nil {
		return dicom.Dataset{}, fmt.Errorf("failed to parse DICOM: %w", err)
	}

	return dicom_dataset, nil
}

func getDicomMetadata(dataset dicom.Dataset) map[string]string {
	metadata := make(map[string]string)

	iterator := dataset.FlatStatefulIterator()
	for iterator.HasNext() {
		element := iterator.Next()
		if element.Value != nil {
			tagInfo, err := tag.Find(element.Tag)
			if err == nil {
				// fmt.Println(tagInfo.Name, tagInfo.Tag.String(), element.Value.String())
				metadata[tagInfo.Name] = element.Value.String()
			}
		}
	}

	return metadata
}

func GetDicomMetadataAsJSON(dataset dicom.Dataset) ([]byte, error) {
	metadata := getDicomMetadata(dataset)
	return json.Marshal(metadata)
}

func GetImageData(dataset dicom.Dataset) ([]byte, error) {
	pixelDataElement, err := dataset.FindElementByTag(tag.PixelData)
	if err != nil {
		return nil, fmt.Errorf("no pixel data found: %w", err)
	}

	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)
	if len(pixelDataInfo.Frames) == 0 {
		return nil, fmt.Errorf("no frames found")
	}

	// Get first frame
	img, err := pixelDataInfo.Frames[0].GetImage()
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

func CreateMetadataHash(dataset dicom.Dataset) string {
	h := sha256.New()
	metadata := getDicomMetadata(dataset)
	fmt.Fprintf(h, "%v", metadata)
	return hex.EncodeToString(h.Sum(nil))
}
