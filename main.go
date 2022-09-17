package main

import (
	"SumoBuzzer2/ShiftOut"
	"machine"
	"time"
)

const SETTING_TIME = 3000
const CHECK_TIME = 200

const (
	SVNSEG_0 = 0x089e
	SVNSEG_1 = 0x0084
	SVNSEG_2 = 0x001f
	SVNSEG_3 = 0x0097
	SVNSEG_4 = 0x0885
	SVNSEG_5 = 0x0893
	SVNSEG_6 = 0x089b
	SVNSEG_7 = 0x0886
	SVNSEG_8 = 0x089f
	SVNSEG_9 = 0x0887
)

type SevenSegDigt struct {
	Digt1 uint8
	Digt2 uint8
	Digt3 uint8
}

var sw1Mode bool
var timer1 uint64
var bz_pattern uint16

func dynamicDrive(ch chan<- uint8) {
	var idx uint8
	idx = 0
	for {
		if idx > 2 {
			idx = 0
		} else {
			idx++
		}
		ch <- idx
		time.Sleep(1 * time.Millisecond)
	}

}
	
// ブザー制御関数
func ctrl_buzer(ch chan<- bool) {
	buz := machine.D6
	buz.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		if ((bz_pattern) & 0x0001) != 0x0000 {
			buz.High()
		} else {
			buz.Low()
		}
		bz_pattern = bz_pattern >> 1
		time.Sleep(100 * time.Millisecond)
		
		ch <- true
	}
}

// 外部割込関数
func pushSw1(sw machine.Pin) {
	sw1Mode = true
}

func main() {
	var outdata uint16
	var svndig SevenSegDigt
	var settingTime int16
	var cntDownStart bool = false
	var stateMode uint8 = 0

	// 74HC595設定
	shout := ShiftOut.New(16, machine.D8, machine.D9, machine.D10)

	// スイッチ設定
	sw1 := machine.D0
	sw1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	sw1.SetInterrupt(machine.PinFalling, pushSw1)
	sw1Mode = false

	//　チャネルの設定
	//ch1 := make(chan bool)
	dd := make(chan uint8)
	bnz := make(chan bool)

	// 表示桁の設定
	svndig.Digt1 = uint8(settingTime / 100)
	svndig.Digt2 = uint8((settingTime % 100) / 10)
	svndig.Digt3 = uint8((settingTime % 10))

	// セグメント変換 無名関数
	NoToSeg := func(no uint8) uint16 {
		var r uint16
		switch no {
		case 0:
			r = SVNSEG_0
		case 1:
			r = SVNSEG_1
		case 2:
			r = SVNSEG_2
		case 3:
			r = SVNSEG_3
		case 4:
			r = SVNSEG_4
		case 5:
			r = SVNSEG_5
		case 6:
			r = SVNSEG_6
		case 7:
			r = SVNSEG_7
		case 8:
			r = SVNSEG_8
		case 9:
			r = SVNSEG_9
		}
		return r
	}

	// スイッチ状態確認 無名関数
	switchCheck := func(t uint64, p machine.Pin) bool {
		var b bool = false
		if p.Get() == true {
			if t >= CHECK_TIME {
				b = true
			}
		}
		return b
	}

	// 変数初期化
	timer1 = 0
	bz_pattern = 0x0000
	settingTime = SETTING_TIME	// 3秒

	// 並行処理ルーチン
	go dynamicDrive(dd) // ダイナミック点灯制御
	go ctrl_buzer(bnz)

	for {

		// ゴルーチン処理
		select {

		// 7セグメントLED表示(ダイナミック点灯制御) 
		// タイマー処理
		case idx := <-dd:
			// タイマーカウント
			timer1++
			
			if cntDownStart {
				if(settingTime % 1000) == 0{
					if settingTime <= 0{
						bz_pattern = 0x03ff
					}else {
						bz_pattern = 0x0001
					}
				}
				settingTime -= 1
				if settingTime < 0 {
					settingTime = 0
					cntDownStart = false
				}
			}

			// 7セグメント制御
			switch idx {
			case 0:
				outdata = NoToSeg(svndig.Digt3)
				outdata |= 0x0100
			case 1:
				outdata = NoToSeg(svndig.Digt2)
				outdata |= 0x0200
			case 2:
				outdata = NoToSeg(svndig.Digt1)
				outdata |= 0x0440
			}
			
			svndig.Digt1 = uint8(settingTime / 1000)
			svndig.Digt2 = uint8((settingTime % 1000) / 100)
			svndig.Digt3 = uint8((settingTime % 100) / 10)

			shout.OutPutData(outdata) // Data Output!
		
		// ブザー
		case _ = <-bnz:
		
		}

		// メインルーチン
		switch stateMode {
		// スイッチ待ち状態
		case 0:
			if sw1Mode {
				cntDownStart = true
				stateMode = 1
				timer1 = 0
				break
			}
		case 1:
			if switchCheck(timer1, sw1) {
				sw1Mode = false
				stateMode = 2
				break
			}
		case 2:
			// フライング
			if sw1Mode {
				cntDownStart = false
				stateMode = 21
				timer1 = 0
				break
			}

			if !cntDownStart {
				stateMode = 11
				timer1 = 0
				break
			}

		case 11:
			if switchCheck(timer1, sw1) {
				sw1Mode = false
				stateMode = 12
				break
			}
		case 12:
			if sw1Mode {
				stateMode = 91
				break
			}

		// フライング処理
		case 21:
			if switchCheck(timer1, sw1) {
				sw1Mode = false
				stateMode = 22
				bz_pattern = 0xaaaa
				break
			}
		case 22:
			// ブザーパターン再設定
			if(bz_pattern & 0xffff) == 0x0000{
				bz_pattern = 0xaaaa
			}

			if sw1Mode {
				stateMode = 23
				timer1 = 0
				bz_pattern = 0x0000
				break
			}
		case 23:
			if switchCheck(timer1, sw1) {
				sw1Mode = false
				stateMode = 24
				break
			}
		case 24:
			if sw1Mode {
				stateMode = 91
				timer1 = 0
				break
			}
			
		// 終了処理
		case 91:
			if switchCheck(timer1, sw1) {
				sw1Mode = false
				stateMode = 92
				break
			}

		case 92:
			settingTime = SETTING_TIME
			stateMode = 0

		}
	}
}
