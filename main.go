package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/urfave/cli"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const (
	ExitCodeOK    int = iota //0
	ExitCodeError int = iota //1
)

func main() {
	err := newApp().Run(os.Args)
	var exitCode = ExitCodeOK
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		exitCode = ExitCodeError
	}
	os.Exit(exitCode)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "tarmd"
	app.HelpName = "tarmd"
	app.Usage = "CLI tool for translate markdown to html and pdf."
	app.UsageText = "tarmd [subcommand] [options]"
	app.Version = "0.1.0"
	app.Author = "lighttiger2505"
	app.Email = "lighttiger2505@gmail.com"
	app.Flags = []cli.Flag{
	// 		cli.StringFlag{
	// 			Name:  "suffix, x",
	// 			Usage: "Diary file suffix",
	// 		},
	}
	app.Commands = cli.Commands{
		{
			Name:    "html",
			Aliases: []string{"h"},
			Usage:   "markdown to html",
			Action:  htmlCommand,
		},
		{
			Name:    "pdf",
			Aliases: []string{"p"},
			Usage:   "markdown to pdf",
			Action:  pdfCommand,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "css, c",
					Usage: "Specific style sheet",
				},
			},
		},
	}
	return app
}

func htmlCommand(c *cli.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return fmt.Errorf("Please input target markdown file")
	}

	mdFilePath := args[0]
	if !isFilePath(mdFilePath) {
		return fmt.Errorf("File not found")
	}

	result, err := toHTML(mdFilePath)
	if err != nil {
		return fmt.Errorf("Failed create html file. %s", err)
	}
	fmt.Println(result)

	return nil
}

func pdfCommand(c *cli.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return fmt.Errorf("Please input target markdown file")
	}

	mdFilePath := args[0]
	if !isFilePath(mdFilePath) {
		return fmt.Errorf("File not found")
	}

	resulthtml, err := toHTML(mdFilePath)
	if err != nil {
		return fmt.Errorf("Failed create html file. %s", err)
	}
	fmt.Println(resulthtml)

	csspath := c.String("css")
	resultpdf, err := toPDF(resulthtml, csspath)
	if err != nil {
		return fmt.Errorf("Failed create pdf file. %s", err)
	}
	fmt.Println(resultpdf)

	return nil
}

func isFilePath(value string) bool {
	absPath, _ := filepath.Abs(value)
	if isFileExist(absPath) {
		return true
	}
	return false
}

func isFileExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return err == nil || !os.IsNotExist(err)
}

func pathToTrimExt(mdFilePath string) string {
	_, filename := filepath.Split(mdFilePath)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func toHTML(mdFilePath string) (string, error) {
	f, err := os.Open(mdFilePath)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	outb := blackfriday.Run(b)

	outname := pathToTrimExt(mdFilePath) + ".html"
	if err := ioutil.WriteFile(outname, outb, 0644); err != nil {
		return "", err
	}
	return outname, nil
}

func toPDF(mdFilePath, cssPath string) (string, error) {
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return "", err
	}

	// Set global options
	// pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	// pdfg.Grayscale.Set(true)

	// Create a new input page from an URL
	page := wkhtmltopdf.NewPage(mdFilePath)

	// Set options for this page
	page.FooterRight.Set("[page]")
	page.FooterFontSize.Set(10)
	// page.Zoom.Set(95.50)

	// Add to document
	pdfg.AddPage(page)

	if cssPath != "" {
		fmt.Println(cssPath)
		pageopt := wkhtmltopdf.NewPageOptions()
		pageopt.UserStyleSheet.Set(cssPath)
		page.PageOptions = pageopt
	}

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		return "", err
	}

	// Write buffer contents to file on disk
	outname := pathToTrimExt(mdFilePath) + ".pdf"
	err = pdfg.WriteFile(outname)
	if err != nil {
		return "", err
	}

	return outname, nil
}
