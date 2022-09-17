package ShiftOut

import (
	"machine"
)

type ShiftOut struct {
	BitSize	int
	ser   machine.Pin // データ入力ピン
	rclk  machine.Pin // ラッチ
	srclk machine.Pin // シフトレジスタクロック
}

// New -> 74HC595 初期化関数
/* (74HC595) Serial Data -> Pin14
   RCLK -> Pin12
   Shift register Clock -> Pin11
*/
func New(bitsize int, ser machine.Pin, rclk machine.Pin, srclk machine.Pin) ShiftOut {

	/* 構造体の宣言 */
	shft := ShiftOut{}

	/* 構造体に格納 */
	shft.ser = ser
	shft.rclk = rclk
	shft.srclk = srclk
	shft.BitSize = bitsize

	/* 各ピンの設定 */
	shft.ser.Configure(machine.PinConfig{Mode: machine.PinOutput})
	shft.rclk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	shft.srclk.Configure(machine.PinConfig{Mode: machine.PinOutput})

	return shft
}

// OutPutData
/* 8bitデータ出力関数 */
/* OutPutData(出力データ) */
func (shft *ShiftOut) OutPutData(outdata uint16) {

	// データ出力
	shft.rclk.Low() // 送信中はラッチをLowに設定

	/* 出力データ設定処理 */
	for i := (shft.BitSize-1); i >=0 ; i-- {

		// 出力データの確認
		if (outdata & (1 << i)) != 0x0000 {
			shft.ser.High() // 1 -> High
		} else {
			shft.ser.Low() // 0 -> Low
		}

		// データ書き込み
		shft.srclk.High()
		shft.srclk.Low()
	}

	shft.rclk.High() // 送信中はラッチをLowに設定
}
