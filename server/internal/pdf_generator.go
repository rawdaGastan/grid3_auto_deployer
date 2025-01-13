package internal

import (
	"fmt"
	"time"

	"github.com/codescalers/cloud4students/models"
	"github.com/pkg/errors"
	"github.com/signintech/gopdf"
)

const (
	greyColor     uint8 = 80
	darkGreyColor uint8 = 60

	startX float64 = 25
	startY float64 = 30

	logoPath       = "internal/img/codescalers.png"
	fontPath       = "internal/fonts/Arial.ttf"
	boldFontPath   = "internal/fonts/Arial-Bold.ttf"
	italicFontPath = "internal/fonts/Arial-Italic.ttf"
)

type InvoicePDF struct {
	invoice models.Invoice
	user    models.User

	pdf    *gopdf.GoPdf
	config gopdf.Config
	startX float64
	startY float64
}

func CreateInvoicePDF(
	invoice models.Invoice, user models.User,
) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	config := gopdf.Config{PageSize: *gopdf.PageSizeA4}
	pdf.Start(config)

	invoicePDF := InvoicePDF{
		invoice: invoice,
		user:    user,
		pdf:     &pdf,
		config:  config,
		startX:  startX,
		startY:  startY,
	}

	if err := invoicePDF.setFonts(); err != nil {
		return nil, errors.Wrap(err, "failed to set fonts")
	}

	return pdf.GetBytesPdf(), errors.Wrap(invoicePDF.draw(), "failed to draw pdf")
}

func (in *InvoicePDF) setFonts() error {
	if err := in.pdf.AddTTFFont("Arial", fontPath); err != nil {
		return err
	}

	if err := in.pdf.AddTTFFont("Arial-Bold", boldFontPath); err != nil {
		return err
	}

	return in.pdf.AddTTFFont("Arial-Italic", italicFontPath)
}

func (in *InvoicePDF) draw() error {
	in.pdf.AddPage()

	if err := in.setLogo(); err != nil {
		return errors.Wrap(err, "failed to set logo")
	}

	// space
	in.startY += 35

	if err := in.title(); err != nil {
		return errors.Wrap(err, "failed to display title")
	}

	// space
	in.startY += 45

	if err := in.companySection(); err != nil {
		return errors.Wrap(err, "failed to display company section")
	}

	if err := in.invoiceSection(); err != nil {
		return errors.Wrap(err, "failed to display invoice section")
	}

	// space
	in.startY += 70

	if err := in.userDetails(); err != nil {
		return errors.Wrap(err, "failed to display user section")
	}

	// space
	in.startY += 90

	if err := in.summary(); err != nil {
		return errors.Wrap(err, "failed to display summary")
	}

	// space
	in.startY += 70

	// Total due
	if err := in.totalDue(); err != nil {
		return errors.Wrap(err, "failed to display total due")
	}

	// space
	in.startY += 85

	// Product usage charges
	if err := in.usageCharges(); err != nil {
		return errors.Wrap(err, "failed to display usage charges")
	}

	// space
	in.startY += 85

	// Table Header
	if err := in.tableHeader(); err != nil {
		return errors.Wrap(err, "failed to display table header")
	}

	// space
	in.startY += 30

	// Table content
	if err := in.tableContent(); err != nil {
		return errors.Wrap(err, "failed to display table content")
	}

	return nil
}

func (in *InvoicePDF) setLogo() error {
	return in.pdf.Image(logoPath, in.startX, in.startY, nil)
}

func (in *InvoicePDF) title() error {
	if err := in.pdf.SetFont("Arial", "", 14); err != nil {
		return err
	}

	in.pdf.SetTextColor(greyColor, greyColor, greyColor)
	in.pdf.SetXY(in.startX, in.startY)

	return in.pdf.Cell(nil,
		fmt.Sprintf("Final invoice for %s %d billing period", in.invoice.CreatedAt.Month().String(), in.invoice.CreatedAt.Year()),
	)
}

func (in *InvoicePDF) companySection() error {
	if err := in.pdf.SetFont("Arial-Bold", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(darkGreyColor, darkGreyColor, darkGreyColor)

	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "From"); err != nil {
		return err
	}

	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(greyColor, greyColor, greyColor)

	in.pdf.SetXY(in.startX, in.startY+15)
	if err := in.pdf.Cell(nil, "Codescalers Egypt"); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX, in.startY+27)
	if err := in.pdf.Cell(nil, "9 Al Wardi street, El Hegaz St"); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX, in.startY+39)
	return in.pdf.Cell(nil, "Cairo Governorate 11341")
}

func (in *InvoicePDF) invoiceSection() error {
	marginRight := float64(250)

	if err := in.pdf.SetFont("Arial-Bold", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(darkGreyColor, darkGreyColor, darkGreyColor)

	in.pdf.SetXY(in.startX+marginRight, in.startY)
	if err := in.pdf.Cell(nil, "Details"); err != nil {
		return err
	}

	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(greyColor, greyColor, greyColor)

	// Data labels
	in.pdf.SetXY(in.startX+250, in.startY+15)
	if err := in.pdf.Cell(nil, "Invoice number:"); err != nil {
		return err
	}
	in.pdf.SetXY(in.startX+250, in.startY+30)
	if err := in.pdf.Cell(nil, "Date of issue:"); err != nil {
		return err
	}
	in.pdf.SetXY(in.startX+250, in.startY+45)
	if err := in.pdf.Cell(nil, "Payment due on:"); err != nil {
		return err
	}

	// Data details
	textWidth, err := in.pdf.MeasureTextWidth(fmt.Sprint(in.invoice.ID))
	if err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-textWidth, in.startY+20)
	if err := in.pdf.Cell(nil, fmt.Sprint(in.invoice.ID)); err != nil {
		return err
	}

	textWidth, err = in.pdf.MeasureTextWidth(in.invoice.CreatedAt.Format("June 1, 2020"))
	if err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-textWidth, in.startY+35)
	if err := in.pdf.Cell(nil, in.invoice.CreatedAt.Format("June 1, 2020")); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-textWidth, in.startY+50)
	return in.pdf.Cell(nil, in.invoice.CreatedAt.Format("June 1, 2020"))
}

func (in *InvoicePDF) userDetails() error {
	if err := in.pdf.SetFont("Arial-Bold", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(darkGreyColor, darkGreyColor, darkGreyColor)

	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "For"); err != nil {
		return err
	}

	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(greyColor, greyColor, greyColor)

	// name
	in.pdf.SetXY(in.startX, in.startY+15)
	if err := in.pdf.Cell(nil, fmt.Sprintf("%s %s", in.user.FirstName, in.user.LastName)); err != nil {
		return err
	}

	// email
	in.pdf.SetXY(in.startX, in.startY+27)
	return in.pdf.Cell(nil, fmt.Sprintf("<%s>", in.user.Email))
}

func (in *InvoicePDF) summary() error {
	if err := in.pdf.SetFont("Arial", "", 14); err != nil {
		return err
	}
	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "Summary"); err != nil {
		return err
	}

	in.pdf.Line(in.startX, in.startY+25, in.startX+540, in.startY+25)
	in.pdf.Line(in.startX, in.startY+55, in.startX+540, in.startY+55)

	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX, in.startY+35)
	if err := in.pdf.Cell(nil, "Total usage charges"); err != nil {
		return err
	}

	totalText := fmt.Sprintf("$%v", in.invoice.Total)
	totalTextWidth, err := in.pdf.MeasureTextWidth(totalText)
	if err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-totalTextWidth, in.startY+35)
	return in.pdf.Cell(nil, fmt.Sprintf("$%v", in.invoice.Total))
}

func (in *InvoicePDF) totalDue() error {
	if err := in.pdf.SetFont("Arial-Bold", "", 14); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "Total due"); err != nil {
		return err
	}

	totalText := fmt.Sprintf("$%v", in.invoice.Total)
	totalTextWidth, err := in.pdf.MeasureTextWidth(totalText)
	if err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-totalTextWidth, in.startY)
	if err := in.pdf.Cell(nil, fmt.Sprintf("$%v", in.invoice.Total)); err != nil {
		return err
	}

	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}
	in.pdf.SetXY(in.startX, in.startY+25)
	if err := in.pdf.Cell(nil, "If you have a credit card on your account, it will be automatically charged within 24 hours"); err != nil {
		return err
	}

	in.pdf.SetStrokeColor(200, 200, 200)
	in.pdf.Line(in.startX, in.startY+60, in.startX+540, in.startY+60)

	return nil
}

func (in *InvoicePDF) usageCharges() error {
	if err := in.pdf.SetFont("Arial", "", 14); err != nil {
		return err
	}
	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "Product usage charges"); err != nil {
		return err
	}

	if err := in.pdf.SetFont("Arial-Italic", "", 10); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX, in.startY+20)
	return in.pdf.Cell(nil, "Detailed usage information ca n be downloaded from the invoices section of your account")
}

func (in *InvoicePDF) tableHeader() error {
	if err := in.pdf.SetFont("Arial-Bold", "", 10); err != nil {
		return err
	}

	in.pdf.SetTextColor(darkGreyColor, darkGreyColor, darkGreyColor)
	in.pdf.SetXY(in.startX, in.startY)
	if err := in.pdf.Cell(nil, "Virtual machines"); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+250, in.startY)
	if err := in.pdf.Cell(nil, "Hours"); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+300, in.startY)
	if err := in.pdf.Cell(nil, "Start"); err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+380, in.startY)
	if err := in.pdf.Cell(nil, "End"); err != nil {
		return err
	}

	totalText := fmt.Sprintf("$%v", in.invoice.Total)
	totalTextWidth, err := in.pdf.MeasureTextWidth(totalText)
	if err != nil {
		return err
	}

	in.pdf.SetXY(in.startX+540-totalTextWidth, in.startY)
	if err := in.pdf.Cell(nil, fmt.Sprintf("$%v", in.invoice.Total)); err != nil {
		return err
	}

	in.pdf.SetStrokeColor(darkGreyColor, darkGreyColor, darkGreyColor)
	in.pdf.Line(in.startX, in.startY+15, in.startX+540, in.startY+15)
	return nil
}

func (in *InvoicePDF) tableContent() error {
	if err := in.pdf.SetFont("Arial", "", 10); err != nil {
		return err
	}
	in.pdf.SetTextColor(greyColor, greyColor, greyColor)

	y := in.startY
	for _, d := range in.invoice.Deployments {
		in.pdf.SetXY(in.startX, y)
		if err := in.pdf.Cell(nil, fmt.Sprintf("vm-%s-%s", d.DeploymentName, d.DeploymentResources)); err != nil {
			return err
		}

		in.pdf.SetXY(in.startX+250, y)
		if err := in.pdf.Cell(nil, fmt.Sprint(d.PeriodInHours)); err != nil {
			return err
		}

		in.pdf.SetXY(in.startX+300, y)
		if err := in.pdf.Cell(nil, d.DeploymentCreatedAt.Format("01-02 15:04")); err != nil {
			return err
		}

		in.pdf.SetXY(in.startX+380, y)
		if err := in.pdf.Cell(nil, time.Now().Format("01-02 15:04")); err != nil {
			return err
		}

		costTextWidth, err := in.pdf.MeasureTextWidth(fmt.Sprintf("$%v", d.Cost))
		if err != nil {
			return err
		}

		in.pdf.SetXY(in.startX+540-costTextWidth, y)
		if err := in.pdf.Cell(nil, fmt.Sprintf("$%v", d.Cost)); err != nil {
			return err
		}

		if y > in.config.PageSize.H-50 {
			in.pdf.AddPage()

			in.startY = startY
			y = in.startY
			if err := in.tableHeader(); err != nil {
				return err
			}

			y += 10
			if err := in.pdf.SetFont("Arial", "", 10); err != nil {
				return err
			}
			in.pdf.SetTextColor(greyColor, greyColor, greyColor)
		}

		y += 15
	}

	return nil
}
