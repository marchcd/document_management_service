package pkg

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/marchcd/kai/internal/models"
	"github.com/nguyenthenguyen/docx"
)

func renderTemplate(tmplPath string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func CreatePDF(ctx context.Context, data *models.DocumentData) ([]byte, error) {
	htmlContent, err := renderTemplate("static/docTempl.html", data)
	if err != nil {
		return nil, fmt.Errorf("template error: %w", err)
	}

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(htmlContent))

	chromeURL := os.Getenv("CHROME_URL")
	if chromeURL == "" {
		chromeURL = "ws://localhost:9222"
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, chromeURL)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte
	err = chromedp.Run(taskCtx,
		chromedp.Navigate(dataURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).WithPaperWidth(8.27).WithPaperHeight(11.69).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp error: %w", err)
	}

	return pdfBuf, nil
}

func CreateRegistryPDF(ctx context.Context, data *models.RegistryData) ([]byte, error) {
	tmpl, err := template.New("registryTempl.html").Funcs(template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).ParseFiles("static/registryTempl.html")
	if err != nil {
		return nil, fmt.Errorf("registry template error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("registry tmpl execute error: %w", err)
	}

	htmlContent := buf.String()
	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(htmlContent))

	chromeURL := os.Getenv("CHROME_URL")
	if chromeURL == "" {
		chromeURL = "ws://localhost:9222"
	}

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, chromeURL)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte
	err = chromedp.Run(taskCtx,
		chromedp.Navigate(dataURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).WithPaperWidth(11.69).WithPaperHeight(8.27).WithLandscape(true).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp registry error: %w", err)
	}

	return pdfBuf, nil
}

func CreateDocx(data *models.DocumentData) ([]byte, error) {
	r, err := docx.ReadDocxFile("static/shemeDoc.docx")
	if err != nil {
		return nil, fmt.Errorf("read docx error: %w", err)
	}
	defer r.Close()

	ed := r.Editable()

	ed.Replace("{{.Issue.Day}}", data.Issue.Day, -1)
	ed.Replace("{{.Issue.Month}}", data.Issue.Month, -1)
	ed.Replace("{{.Issue.YearShort}}", data.Issue.YearShort, -1)
	ed.Replace("{{.Issue.Number}}", data.Issue.Number, -1)

	employerLine1, employerLine2 := splitStringName(data.Employer.Name, 78)
	ed.Replace("{{.Employer.Name}}", employerLine1, -1)
	ed.Replace("{{.Employer.NameExt}}", employerLine2, -1)

	ed.Replace("{{.Student.FullTitle}}", data.Student.FullTitle, -1)
	ed.Replace("{{.Student.EducationForm}}", data.Student.EducationForm, -1)
	ed.Replace("{{.Student.Course}}", fmt.Sprintf("%d", data.Student.Course), -1)

	ed.Replace("{{.Period.StartDay}}", data.Period.StartDay, -1)
	ed.Replace("{{.Period.StartMonth}}", data.Period.StartMonth, -1)
	ed.Replace("{{.Period.StartYear}}", data.Period.StartYear, -1)
	ed.Replace("{{.Period.EndDay}}", data.Period.EndDay, -1)
	ed.Replace("{{.Period.EndMonth}}", data.Period.EndMonth, -1)
	ed.Replace("{{.Period.EndYear}}", data.Period.EndYear, -1)
	ed.Replace("{{.Period.Duration}}", fmt.Sprintf("%d", data.Period.Duration), -1)

	specialtyLine1, specialtyLine2 := splitStringName(data.Student.Specialty, 41)
	ed.Replace("{{.Student.Specialty}}", specialtyLine1, -1)
	ed.Replace("{{.Student.SpecialtyExt}}", specialtyLine2, -1)

	ed.Replace("{{.Student.FullTitleNominative}}", data.Student.FullTitleNominative, -1)

	ed.Replace("{{.Period.StartDay}}", data.Period.StartDay, -1)
	ed.Replace("{{.Period.StartMonth}}", data.Period.StartMonth, -1)
	ed.Replace("{{.Period.StartYear}}", data.Period.StartYear, -1)
	ed.Replace("{{.Period.EndDay}}", data.Period.EndDay, -1)
	ed.Replace("{{.Period.EndMonth}}", data.Period.EndMonth, -1)
	ed.Replace("{{.Period.EndYear}}", data.Period.EndYear, -1)
	ed.Replace("{{.Period.Duration}}", fmt.Sprintf("%d", data.Period.Duration), -1)

	var out bytes.Buffer
	err = ed.Write(&out)
	if err != nil {
		return nil, fmt.Errorf("write docx error: %w", err)
	}

	return out.Bytes(), nil
}

func splitStringName(name string, limit int) (string, string) {
	runes := []rune(name)
	if len(runes) <= limit {
		return name, ""
	}

	splitIdx := limit
	for i := limit; i > 0; i-- {
		if runes[i] == ' ' {
			splitIdx = i
			break
		}
	}

	return string(runes[:splitIdx]), string(runes[splitIdx:])
}
