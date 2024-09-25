package main

import (
	"encoding/xml"
	"image"
	"image/jpeg"
	"log"
	"os"

	"github.com/fogleman/gg"
)

const (
	configFile = "config.xml"
	input      = "inputs/input.jpg"
	output     = "outputs/output.jpg"
	fontFolder = "font/"
)

var (
	imgWidth     int
	imgHeight    int
	totalHeight  int
	bar          Bar
	imageObject  image.Image
	barXPosition int
	barYPosition int
)

type Bar struct {
	Position string    `xml:"position,attr"`
	Color    string    `xml:"color,attr"`
	Height   int       `xml:"height,attr"`
	Sections []Section `xml:"section"`
}

type Section struct {
	Height int     `xml:"height,attr"`
	Font   Font    `xml:"font"`
	Panels []Panel `xml:"panel"`
}

type Font struct {
	Name  string `xml:"name,attr"`
	Color string `xml:"color,attr"`
	Size  int    `xml:"size,attr"`
}

type Panels struct {
	Rows    int     `xml:"rows,attr"`
	Columns int     `xml:"columns,attr"`
	Panel   []Panel `xml:"panel"`
}

type Panel struct {
	Color       string  `xml:"color,attr"`
	BorderWidth int     `xml:"borderWidth,attr"`
	BorderColor string  `xml:"borderColor,attr"`
	Labels      []Label `xml:"label"`
	Values      []Label `xml:"value"`
}

type Label struct {
	Text  string `xml:"text,attr"`
	Align string `xml:"align,attr"`
}

func loadXML() Bar {
	xmlFile, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to open XML file: %v", err)
	}
	defer xmlFile.Close()

	var bar Bar
	decoder := xml.NewDecoder(xmlFile)
	if err := decoder.Decode(&bar); err != nil {
		log.Fatalf("Failed to decode XML file: %v", err)
	}
	return bar
}

func readImgInfo() {
	imgFile, err := os.Open(input)
	if err != nil {
		log.Fatalf("Failed to open input image file: %v", err)
	}
	defer imgFile.Close()

	img, err := jpeg.Decode(imgFile)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	imageObject = img

	imgBounds := imageObject.Bounds()
	imgWidth = imgBounds.Dx()
	imgHeight = imgBounds.Dy()
	totalHeight = imgHeight + bar.Height

	// bar position
	barXPosition = 0
	barYPosition = imgHeight
}

func drawBox(dc *gg.Context, sectionIndex int, panelIndex int, labelIndex int) {

	yPosition := 0
	for i := 0; i < sectionIndex; i++ {
		yPosition += bar.Sections[i].Height
	}
	yPosition += bar.Sections[sectionIndex].Height /
		len(bar.Sections[sectionIndex].Panels[panelIndex].Labels) *
		labelIndex

	xPosition := imgWidth /
		len(bar.Sections[sectionIndex].Panels) * panelIndex

	width := imgWidth / len(bar.Sections[sectionIndex].Panels)
	height := bar.Sections[sectionIndex].Height /
		len(bar.Sections[sectionIndex].Panels[panelIndex].Labels)

	// label and value pos regarding alignment
	labelXPos := 0
	valueXPos := 0
	labelAlign := 0
	valueAlign := 0
	if bar.Sections[sectionIndex].Panels[panelIndex].Labels[labelIndex].Align == "L" {
		labelXPos = 15
		labelAlign = 0
	} else if bar.Sections[sectionIndex].Panels[panelIndex].Labels[labelIndex].Align == "R" {
		labelXPos = width / 2
		labelAlign = 1
	}
	if len(bar.Sections[sectionIndex].Panels[panelIndex].Values) > labelIndex {
		if bar.Sections[sectionIndex].Panels[panelIndex].Values[labelIndex].Align == "L" {
			valueXPos = width/2 + 15
			valueAlign = 0
		} else if bar.Sections[sectionIndex].Panels[panelIndex].Values[labelIndex].Align == "R" {
			valueXPos = width - 15
			valueAlign = 1
		}
	}

	// draw box
	dc.SetHexColor(bar.Sections[sectionIndex].Panels[panelIndex].Color)
	dc.DrawRectangle(
		float64(barXPosition+xPosition),
		float64(barYPosition+yPosition),
		float64(width),
		float64(height),
	)
	dc.Fill()

	// draw text
	if err := dc.LoadFontFace(
		fontFolder+bar.Sections[sectionIndex].Font.Name,
		float64(bar.Sections[sectionIndex].Font.Size),
	); err != nil {
		log.Fatalf("Failed to load font: %v", err)
	}

	dc.SetHexColor(bar.Sections[sectionIndex].Font.Color)
	dc.DrawStringAnchored(
		bar.Sections[sectionIndex].Panels[panelIndex].Labels[labelIndex].Text+":",
		float64(barXPosition+xPosition+labelXPos),
		float64(barYPosition+yPosition+height/2),
		float64(labelAlign),
		0.5,
	)

	if len(bar.Sections[sectionIndex].Panels[panelIndex].Values) > labelIndex {
		dc.DrawStringAnchored(
			bar.Sections[sectionIndex].Panels[panelIndex].Values[labelIndex].Text,
			float64(barXPosition+xPosition+valueXPos),
			float64(barYPosition+yPosition+height/2),
			float64(valueAlign),
			0.5,
		)
	}

	// draw border
	dc.SetHexColor(bar.Sections[sectionIndex].Panels[panelIndex].BorderColor)
	dc.SetLineWidth(float64(bar.Sections[sectionIndex].Panels[panelIndex].BorderWidth))
	dc.DrawRectangle(
		float64(barXPosition+xPosition),
		float64(barYPosition+yPosition),
		float64(width),
		float64(height),
	)
	dc.Stroke()

}

func draw(dc *gg.Context) {

	// set bar background
	dc.SetHexColor(bar.Color)
	dc.Clear()

	// draw section
	for i := 0; i < len(bar.Sections); i++ {
		for j := 0; j < len(bar.Sections[i].Panels); j++ {
			for k := 0; k < len(bar.Sections[i].Panels[j].Labels); k++ {
				drawBox(dc, i, j, k)
			}

		}
	}
}

func generateImage(dc *gg.Context) {
	outFile, err := os.Create(output)
	if err != nil {
		log.Fatalf("Failed to encode output image: %v", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dc.Image(), nil); err != nil {
		log.Fatalf("Failed to encode output image: %v", err)
	}
}

func main() {
	bar = loadXML()

	// read img info
	readImgInfo()

	// create draw object
	dc := gg.NewContext(imgWidth, totalHeight)

	// draw img
	draw(dc)
	dc.DrawImage(imageObject, 0, 0)

	// generate image
	generateImage(dc)

}
