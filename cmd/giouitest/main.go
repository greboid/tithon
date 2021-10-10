package main

import (
	"fmt"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("IRC Client"),
			app.Size(unit.Sp(800), unit.Sp(600)),
		)
		var ops op.Ops
		var textPane richtext.InteractiveText
		var inputField = widget.Editor{
			SingleLine: true,
			Submit:     true,
		}
		var serverList = widget.List{
			Scrollbar: widget.Scrollbar{},
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: false,
				Position:    layout.Position{},
			},
		}
		var nickList = widget.List{
			Scrollbar: widget.Scrollbar{},
			List: layout.List{
				Axis:        layout.Vertical,
				ScrollToEnd: false,
				Alignment:   0,
				Position:    layout.Position{},
			},
		}
		var spans = []richtext.SpanStyle{
			{
				Content: fmt.Sprintf(
					"%s <%s> %s\r\n",
					time.Now().Format("15:04:05"),
					"Greboid",
					"Oh god, I hate textpanes so much.  Please kill me.",
				),
				Color: color.NRGBA{A: 255},
				Size:  unit.Dp(12),
				Font:  gofont.Collection()[0].Font,
			},
		}
		th := material.NewTheme(gofont.Collection())
		for e := range w.Events() {
			switch e := e.(type) {
			case system.DestroyEvent:
				os.Exit(0)
			case system.FrameEvent:
				for _, e := range inputField.Events() {
					if _, ok := e.(widget.SubmitEvent); ok {
						text := inputField.Text()
						if len(text) != 0 {
							inputField.SetText("")
							spans = append(spans, richtext.SpanStyle{
								Content: fmt.Sprintf(
									"%s <%s> %s\r\n",
									time.Now().Format("15:04:05"),
									"User",
									text,
								),
								Color: color.NRGBA{A: 255},
								Size:  unit.Dp(12),
								Font:  gofont.Collection()[0].Font,
							})
						}
					}
				}
				gtx := layout.NewContext(&ops, e)
				layout.UniformInset(unit.Sp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceStart,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							list := material.List(th, &serverList)
							return list.Layout(gtx, 8, func(gtx layout.Context, i int) layout.Dimensions {
								if i == 0 || i == 6 {
									return material.Label(
										th,
										unit.Sp(12),
										fmt.Sprintf("Server %d", i),
									).Layout(gtx)
								} else {
									return layout.Inset{
										Left: unit.Sp(10),
									}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return material.Label(
											th,
											unit.Sp(12),
											fmt.Sprintf("Channel %d", i),
										).Layout(gtx)
									},
									)
								}
							})
						}),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis:    layout.Vertical,
								Spacing: layout.SpaceStart,
							}.Layout(gtx,
								layout.Flexed(100, func(gtx layout.Context) layout.Dimensions {
									return richtext.Text(&textPane, th.Shaper, spans...).Layout(gtx)
								}),
								layout.Rigid(
									layout.Spacer{Height: unit.Sp(4)}.Layout,
								),
								layout.Rigid(
									func(gtx layout.Context) layout.Dimensions {
										btn := material.Editor(th, &inputField, "")
										return btn.Layout(gtx)
									},
								),
							)
						}),
						layout.Rigid(
							layout.Spacer{Width: unit.Sp(5)}.Layout,
						),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							list := material.List(th, &nickList)
							return list.Layout(gtx, 10, func(gtx layout.Context, i int) layout.Dimensions {
								return material.Label(
									th,
									unit.Sp(12),
									fmt.Sprintf("Nickname %d", i),
								).Layout(gtx)
							})
						}),
						layout.Rigid(
							layout.Spacer{Width: unit.Sp(5)}.Layout,
						),
					)
				})
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
