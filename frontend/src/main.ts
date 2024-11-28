import { createChart,CrosshairMode,CandlestickData,LineData,SeriesMarker} from 'lightweight-charts';
import {Quote,Indicator} from './types';
import { convTime } from './help.js';


export function buildSMA(quote: Quote,sma:Indicator[]): void {

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
	const lineSeries = chart.addLineSeries({});  

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

	let lineSeriesData: LineData[] = [];
	let markersChart: any[]= [];
	for (let item of sma) {
		if (item.value!=0){
			lineSeriesData.push({time: convTime(item.date),value:item.value});
			markersChart.push({ time: convTime(item.date), position: 'inBar', color: '#0000FF', shape: 'circle' });  
		}  
	}
	lineSeries.setData(lineSeriesData);
}






