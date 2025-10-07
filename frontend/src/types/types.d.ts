export interface Quote {
	symbol: string;
	precision: number;
	date: string[]; // Преобразуем time.Time в строку в формате ISO
	open: number[];
	high: number[];
	low: number[];
	close: number[];
	volume: number[];
}


export interface Indicator {
    date: string; // Преобразуем time.Time в строку в формате ISO
    value: number;
}


type SniperConfig  = {
  RSILength: number;
  EMAFastLength: number;
  EMASlowLength: number;
  ATRLength: number;
  ATRSMALength: number;
  VolumeSMALength: number;
  RSIMFIBuyLevel: number;
  RSIMFIExitLevel: number;
  EMAMinDelta: number;
  ATRMultiplier: number;
  BuyVolumeFactor: number;
  SellVolumeFactor: number;
};


const sniperConfig: SniperConfig = {
  RSILength: 14,
  EMAFastLength: 8,
  EMASlowLength: 50,
  ATRLength: 14,
  ATRSMALength: 14,
  VolumeSMALength: 20,
  RSIMFIBuyLevel: 38.0,
  RSIMFIExitLevel: 50.0, // 62
  EMAMinDelta: 0.0001,
  ATRMultiplier: 0.5,
  BuyVolumeFactor: 1.2,
  SellVolumeFactor: 0.8
};