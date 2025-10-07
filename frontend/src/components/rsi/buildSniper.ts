import { createChart, CrosshairMode } from 'lightweight-charts';
import type { CandlestickData, LineData } from 'lightweight-charts';

import type  {Quote,Indicator} from '../../types/types.d';




interface SniperChartInstance {
  update: (quote: Quote, signalBuyPoints: Indicator[], signalSellPoints: Indicator[]) => void
  destroy: () => void
}


export function buildSniper(
  quote: Quote,
  signalBuyPoints: Indicator[],
  signalSellPoints: Indicator[],
  container?: HTMLElement
): SniperChartInstance {

  const container_chart = container ?? document.getElementById('chart')!
  container_chart.innerHTML = '' // очищаем контейнер, если пересоздаём

  // Настройки графика
  const chartOptions = {
    height: 500,
    width: container_chart.clientWidth,
    autosize: true,
    layout: {
      backgroundColor: '#ffffff',
      textColor: 'rgba(33, 56, 77, 1)',
    },
    grid: {
      vertLines: { color: 'rgba(197, 203, 206, 0.7)' },
      horzLines: { color: 'rgba(197, 203, 206, 0.7)' },
    },
    crosshair: { mode: CrosshairMode.Normal },
    timeScale: {
      timeVisible: true,
      secondsVisible: false,
    },
  }

  const chart = createChart(container_chart, chartOptions)
  const candleSeries = chart.addCandlestickSeries({})
  const lineSeriesHigh = chart.addLineSeries({
    color: 'rgba(255, 255, 255, 0)',
    lastValueVisible: false,
    priceLineVisible: false,
  })
  const lineSeriesLow = chart.addLineSeries({
    color: 'rgba(255, 255, 255, 0)',
    lastValueVisible: false,
    priceLineVisible: false,
  })

  // Функция обновления данных графика
  const updateChart = (quote: Quote, signalBuyPoints: Indicator[], signalSellPoints: Indicator[]) => {
    if (!quote || !quote.date?.length) return

    // --- свечи ---
    const candles: CandlestickData[] = []
    for (let i = 0; i < quote.close.length; i++) {
      candles.push({
        time: convTime(quote.date[i]),
        open: quote.open[i],
        high: quote.high[i],
        low: quote.low[i],
        close: quote.close[i],
      })
    }
    candleSeries.setData(candles)

    // --- сигналы ---
    const lineSeriesDataHigh: LineData[] = []
    const lineSeriesDataLow: LineData[] = []
    const markersChartHigh: any[] = []
    const markersChartLow: any[] = []

    for (const item of signalBuyPoints) {
      if (item.value !== 0) {
        lineSeriesDataHigh.push({ time: convTime(item.date), value: item.value })
        markersChartHigh.push({ time: convTime(item.date), position: 'belowBar', color: '#008000', shape: 'arrowUp' })
      }
    }

    for (const item of signalSellPoints) {
      if (item.value !== 0) {
        lineSeriesDataLow.push({ time: convTime(item.date), value: item.value })
        markersChartLow.push({ time: convTime(item.date), position: 'aboveBar', color: '#FF0000', shape: 'arrowDown' })
      }
    }

    lineSeriesHigh.setData(lineSeriesDataHigh)
    lineSeriesLow.setData(lineSeriesDataLow)
    lineSeriesHigh.setMarkers(markersChartHigh)
    lineSeriesLow.setMarkers(markersChartLow)
  }

  // Первичное построение
  updateChart(quote, signalBuyPoints, signalSellPoints)

  // Возвращаем экземпляр для реактивного обновления
  return {
    update: (newQuote, newBuy, newSell) => {
      updateChart(newQuote, newBuy, newSell)
    },
    destroy: () => {
      chart.remove()
    },
  }
}



export function buildSniperOld(quote: Quote,signalBuyPoints:Indicator[],signalSellPoints:Indicator[]): void {

	const container_chart = document.getElementById('chart')!;
	const chartOptions = {
		height: 500,
		width: 700,
		autosize: true,
		layout: {
			backgroundColor: '#ffffff',
			textColor: 'rgba(33, 56, 77, 1)',
		},
		grid: {
			vertLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
			horzLines: {
				color: 'rgba(197, 203, 206, 0.7)',
			},
		},
		crosshair: {
			mode: CrosshairMode.Normal,
		},
		timeScale: {
			timeVisible: true,
			secondsVisible: false
		},
	};

	const chart = createChart(container_chart, chartOptions);
	const candleSeries = chart.addCandlestickSeries({});
	const lineSeriesHigh = chart.addLineSeries({
		color: 'rgba(255, 255, 255, 0)', 
		lastValueVisible: false, 
		priceLineVisible: false
	}); 
	const lineSeriesLow = chart.addLineSeries({
		color: 'rgba(255, 255, 255, 0)', 
		lastValueVisible: false, 
		priceLineVisible: false
	});

	let candles:CandlestickData[] = [];
	for (let index = 0; index < quote.close.length; index++) {
		candles.push({ 
			time: convTime(quote.date[index]), 
			open: quote.open[index],
			high: quote.high[index],
			low: quote.low[index], 
			close: quote.close[index] 
			});
	}
	candleSeries.setData(candles);

	let lineSeriesDataHigh: LineData[] = [];
	let lineSeriesDataLow: LineData[] = [];
	let markersChartHigh: any[]= [];
	let markersChartLow: any[]= [];
	for (let item of signalBuyPoints) {
		if (item.value!=0){
			lineSeriesDataHigh.push({time: convTime(item.date),value:item.value});
			// markersChartHigh.push({ time: convTime(item.date), position: 'inBar', color: '#008000', shape: 'circle' }); 
			// buy
			markersChartHigh.push({ time: convTime(item.date), position: 'belowBar', color: '#008000', shape: 'arrowUp' }); 
		}  
	}

	for (let item of signalSellPoints) {
		if (item.value!=0){
			lineSeriesDataLow.push({time: convTime(item.date),value:item.value});
			markersChartLow.push({ time: convTime(item.date), position: 'inBar', color: '#FF0000', shape: 'circle' }); 
			// sell
			markersChartLow.push({ time: convTime(item.date), position: 'aboveBar', color: '#FF0000', shape: 'arrowDown' }); 
		}  
	}

	lineSeriesHigh.setData(lineSeriesDataHigh);
	lineSeriesLow.setData(lineSeriesDataLow);
	lineSeriesHigh.setMarkers(markersChartHigh);
	lineSeriesLow.setMarkers(markersChartLow);
}


function convTime(time:string):any {
    const date = new Date(time);
    return Math.floor(date.getTime() / 1000);  // Преобразование в Unix timestamp в секунды 
}