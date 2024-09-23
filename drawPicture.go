package main

import (
	"bytes"
	"crypto/rand"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
)

func makePictureNotFound() []byte {
	w := 1024
	h := 576
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	black := color.RGBA{10, 10, 10, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)
	fontBytes, _ := staticFiles.ReadFile("build/arial.ttf")
	font, _ := freetype.ParseFont(fontBytes)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(48)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)
	pt := freetype.Pt(400, 240+int(c.PointToFixed(48)>>6))
	_, _ = c.DrawString("Not found!", pt)

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	imageData, err := io.ReadAll(&buf)
	if err != nil {
		panic(err)
	}
	return imageData

}

func makePictureCh(id string) bool {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 576))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.Black}, image.Point{}, draw.Src)
	fontBytes, _ := staticFiles.ReadFile("build/arial.ttf")
	font, _ := freetype.ParseFont(fontBytes)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(36)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)
	x := 430
	if strToInt(id) > 9 {
		x -= 20
	}
	y := 40
	c.DrawString("Archive #"+id, freetype.Pt(x, y+int(c.PointToFixed(36)>>6)))

	c.SetFontSize(18)
	x = 265
	y = 390
	c.DrawString("This channel is intended for watching recorded TV programs", freetype.Pt(x, y+int(c.PointToFixed(36)>>6)))

	tm := conf("trademark")
	rnd, _ := rand.Int(rand.Reader, big.NewInt(9))
	if rnd.Int64() == int64(0) {
		x = 455
		tm = "TVHOST.CC"
	}
	y = 420
	c.SetSrc(image.NewUniform(color.RGBA{81, 207, 236, 255}))
	c.DrawString(tm, freetype.Pt(x, y+int(c.PointToFixed(36)>>6)))

	//добавляем qr
	tmpLink = md5hash(toStr(time.Now().UnixNano()))
	qrFile := tmpDir + "/qr.png"
	qrcode.WriteFile(conf("site")+"/?ch="+tmpLink, qrcode.Medium, 256, qrFile)
	overlayFile, _ := os.Open(qrFile)
	overlayQr, _ := png.Decode(overlayFile)
	x = 380
	y = 120
	draw.Draw(img, image.Rect(x, y, x+overlayQr.Bounds().Dx(), y+overlayQr.Bounds().Dy()), overlayQr, image.Point{}, draw.Over)
	os.Remove(tmpDir + "/qr.png")

	// напишем время обновления
	c.SetSrc(image.NewUniform(color.RGBA{255, 0, 0, 255}))
	c.DrawString("Upd: "+time.Now().Format("15:04"), freetype.Pt(20, 10+int(c.PointToFixed(36)>>6)))

	// сохраняем файл backkground
	totalJpg := tmpDir + "/bg" + id + ".jpg"
	file, _ := os.Create(totalJpg)
	defer file.Close()
	jpeg.Encode(file, img, &jpeg.Options{Quality: 100})

	// копируем овердохрена картинок
	os.RemoveAll(tmpDir + "/" + id)
	chkDir(tmpDir + "/" + id)
	for i := 0; i < 500; i++ {
		copyFile(totalJpg, tmpDir+"/"+id+"/img"+toStr(i)+".jpg")
	}

	return true

}

func makePictureCh0() {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 576))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.Black}, image.Point{}, draw.Src)
	fontBytes, _ := staticFiles.ReadFile("build/arial.ttf")
	font, _ := freetype.ParseFont(fontBytes)
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(36)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	c.DrawString("Archive channels", freetype.Pt(370, 20+int(c.PointToFixed(36)>>6)))

	// перечисляем потоки
	r := db_query("SELECT * FROM streams ORDER BY id")

	fontContentSize := 21
	c.SetFontSize(float64(fontContentSize))
	for k, v := range r {

		vv := v.(map[string]string)

		c.SetSrc(image.White) // white
		id := strToInt(conf("lcnmain")) + strToInt(vv["id"])
		xx := 60
		if id > 99 {
			xx = 40
		} else if id > 9 {
			xx = 50
		}

		yy := 105 + k*(fontContentSize+2)
		c.DrawString(toStr(id)+".", freetype.Pt(xx, yy))

		if vv["enable"] == "0" {
			c.SetSrc(image.NewUniform(color.RGBA{100, 100, 100, 255})) // gray
			c.DrawString("OFF", freetype.Pt(90, yy))
		} else if vv["play"] == "" {
			c.SetSrc(image.NewUniform(color.RGBA{0, 255, 0, 255})) // green
			c.DrawString("FREE", freetype.Pt(90, yy))
		} else {
			spl := strings.Split(vv["play"], ":")
			//
			//
			//
			/////////////////////////
			// налаживаем логотип
			logoFile := "./logos/" + spl[0] + ".png"
			pngFile, err := os.Open(logoFile)
			if err == nil {
				defer pngFile.Close()
				pngImg, err := png.Decode(pngFile)
				if err != nil {
					panic(err)
				}
				resizedPngImg := resize.Resize(22, 22, pngImg, resize.Lanczos3)
				offset := image.Pt(90, yy-19)
				b := resizedPngImg.Bounds()
				draw.Draw(img, b.Add(offset), resizedPngImg, image.Point{}, draw.Over)
			}
			////////////////////////
			//
			//
			//
			//

			rr := db_fetchassoc(db_query("SELECT * FROM epg WHERE id=" + spl[1]))
			s_time := int64(strToInt(vv["b_time"]))
			c.DrawString(time.Unix(s_time, 0).Format("15:04"), freetype.Pt(126, yy))
			c.DrawString("-", freetype.Pt(182, yy))
			f_time := s_time + int64(strToInt(rr["f_time"])-strToInt(rr["s_time"]))
			c.DrawString(time.Unix(f_time, 0).Format("15:04"), freetype.Pt(192, yy))

			c.SetSrc(image.White) // white
			c.DrawString(rr["title"]+" ("+time.Unix(int64(strToInt(rr["s_time"])), 0).Format("Mon, 02 Jan 15:04")+")", freetype.Pt(260, yy))
		}
	}

	c.SetFontSize(18)

	// красный цвет UPD: 15:04
	c.SetSrc(image.NewUniform(color.RGBA{255, 0, 0, 255}))
	c.DrawString("Upd: "+time.Now().Format("15:04"), freetype.Pt(20, 10+int(c.PointToFixed(18)>>6)))

	// сохраняем файл backkground
	totalJpg := tmpDir + "/bg0.jpg"
	file, _ := os.Create(totalJpg)
	defer file.Close()
	jpeg.Encode(file, img, &jpeg.Options{Quality: 100})

	// копируем овердохрена картинок
	os.RemoveAll(tmpDir + "/0")
	chkDir(tmpDir + "/0")
	for i := 0; i < 450; i++ {
		copyFile(totalJpg, tmpDir+"/0/img"+toStr(i)+".jpg")
	}
}
